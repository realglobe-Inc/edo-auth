-- Copyright 2015 realglobe, Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.


-- クライアント認証型 TA 認証だけのプロキシ。

local openssl = require("openssl")
local varutil = require("lib.varutil")
local erro = require("lib.erro")
local redis_wrapper = require("lib.redis_wrapper")
local key_db = require("lib.key_db")
local session = require("lib.session")
local session_db = require("lib.session_db")
local pre_session = require("lib.pre_session")
local pre_session_db = require("lib.pre_session_db")


local function get_absolute_path(path)
   if not path then
      -- 知らん。
      return nil
   elseif string.len(path) > 0 and string.sub(path, 1, 1) == "/" then
      -- 絶対パス。
      return path
   else
      -- 相対パス。
      return ngx.config.prefix() .. "/" .. path
   end
end


-- $edo_log_level: デバッグログのレベル。
local log_level = varutil.get_level(ngx.var.edo_log_level)
-- $edo_session_length: セッション ID の文字列長。
local session_length = ngx.var.edo_session_length or 30
-- $edo_token_length: 署名用文字列の長さ。
local token_length = ngx.var.edo_token_length or 30
-- $edo_default_hash: 無指定時の署名検証に使われるハッシュ関数。
local default_hash = ngx.var.edo_default_hash or "sha256"
-- $edo_start_delay: セッションの開始から認証完了までの猶予期間 (秒)。
local start_delay = ngx.var.edo_start_delay or 3 * 60 -- 3 分。
-- $edo_redis_address: redis のアドレス。
local redis_address = ngx.var.edo_redis_address or "127.0.0.1:6379"
-- $edo_redis_timeout: redis の接続待ち時間 (ミリ秒)。
local redis_timeout = ngx.var.edo_redis_timeout or 30 * 1000 -- 30 秒。
-- $edo_redis_keepalive: redis ソケットの待機時間 (ミリ秒)。
local redis_keepalive = ngx.var.edo_redis_keepalive or 60 * 1000 -- 1 分。
-- $edo_redis_pool_size: 1 ワーカー当たりの待機させる redis 接続数。
local redis_pool_size = ngx.var.edo_redis_pool_size or 16
-- $edo_redis_session_tag: 確立済みセッションを redis に格納する際のキーの接頭辞。
local redis_session_tag = ngx.var.edo_redis_session_tag or "session"
-- $edo_redis_pre_session_tag: 未確立セッションを redis に格納する際のキーの接頭辞。
local redis_pre_session_tag = ngx.var.edo_redis_pre_session_tag or "pre-session"
-- $edo_key_directory: 公開鍵が <TA の ID>.pub って名前で入ってるディレクトリ。
local key_directory = get_absolute_path(ngx.var.edo_key_directory)
-- $edo_redis_key_tag: 公開鍵を redis に格納する際のキーの接頭辞。
local redis_key_tag = ngx.var.edo_redis_key_tag or "key"
-- $edo_session_expires_in: セッションの有効期間 (秒)。
local session_expires_in = ngx.var.edo_session_expires_in or 60 * 60 -- 1 時間。
-- $edo_cache_expires_in: キャッシュの有効期間 (秒)。
local cache_expires_in = ngx.var.edo_cache_expires_in or 10 * 60 -- 10 分。
-- $edo_cookie_path: cookie の Path。複数 TA を混在させるならあった方が良いか？
local cookie_path = ngx.var.edo_cookie_path or "/"


-- エラーレスポンスを返す。
local function respond_json(params)
   if params.message then
      ngx.header["X-Edo-Auth-Ta-Error"] = params.message
   else
      ngx.header["X-Edo-Auth-Ta-Error"] = "error occurred"
   end
   return erro.respond_json(params)
end


-- ランダムな文字列を生成する。
local function generate_random_string(length)
   -- OpenSSL の疑似乱数列を BASE64 エンコードする。
   local buff, err = openssl.random((length * 3  + 3) / 4, false)
   if err then
      return nil, {message = "openssl.random failed"}
   end
   local s = ngx.encode_base64(buff)
   if string.len(s) > length then -- BASE64 は末尾に = を入れる場合があるから。
      return string.sub(s, 1, length)
   else
      return s
   end
end


-- リクエスト元のアドレスを特定する。
-- 手前にリバースプロキシが居たら、X-Forwarded-For ヘッダフィールドを見る。
-- クライアントが設定した X-Forwarded-For ヘッダフィールドを通さないように、
-- nginx や手前のリバースプロキシの設定を書く必要あり。
local function get_request_address()
   local forwarded = ngx.var.http_x_forwarded_for
   if forwarded then
      -- X-Forwarded-For ヘッダフィールドがある。
      ngx.log(log_level, "use X-Forwarded-For to get request source")

      local idx = string.find(forwarded, ",")
      if idx then
         return string.sub(forwarded, 1, idx)
      else
         return forwarded
      end
   end
   return ngx.var.remote_addr
end


-- レスポンスに Set-Cookie ヘッダをつける。
local function set_cookie(name, value, options)
   -- ブラウザ向けではないので、必要なものだけで、Secure や HttpOnly は要らない。
   local line = name.."="..value
   for key, val in pairs(options) do
      line = line.."; "..key.."="..val
   end
   ngx.header["Set-Cookie"] = line
end


-- リクエストの Cookie から値を除く。
local function filter_cookie(name, value)
   -- BASE64 の + を \+ にしてから正規表現として ngx.re.sub に渡す。
   local v, _, err = ngx.re.gsub(value, "\\+", "\\+")
   if err then
      return {message = "session filter preprocessing failed"}
   end
   local after, _, err = ngx.re.sub(ngx.var.http_cookie, name .. " *= *" .. v .. ";?", "", "i")
   if err then
      return {message = "session filtering failed"}
   end
   ngx.req.set_header("Cookie", after)
end


-- 署名検証。
local function verify(token, sign, public_key_pem, hash)
   local raw_sign = ngx.decode_base64(sign)
   if not raw_sign then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "bad sign format"}
   end

   local public_key, err = openssl.pkey.read(public_key_pem)
   if err then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = err}
   end
   return public_key:verify(token, raw_sign, hash)
end


-- セッションを開始する。
local function start_session(pre_sess_db)
   -- 乱数列、セッション ID を生成し、
   -- セッション ID にリクエスト元アドレスと乱数列を紐付けて保存し、
   -- 401 Unauthorized でセッション ID と乱数列を返す。

   local address = get_request_address()
   if not address then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "cannot get unauthenticated request source"})
   end

   -- リクエスト元のアドレスを特定した。
   ngx.log(log_level, "unauthenticated request source is " .. address)

   local session_id, err = generate_random_string(session_length)
   if err then
      return respond_json(err)
   end

   -- セッション ID を生成した。
   ngx.log(log_level, "session id was generated")

   local token, err = generate_random_string(token_length)
   if err then
      return respond_json(err)
   end

   -- 署名用乱数列を生成した。
   ngx.log(log_level, "token was generated")

   local pre_sess = pre_session.new(session_id, token, address)
   local err = pre_sess_db:save(pre_sess, start_delay)
   if err then
      return respond_json(err)
   end

   -- セッションを登録した。
   ngx.log(log_level, "session began")

   set_cookie("X-Edo-Auth-Ta-Session", pre_sess:get_id(), {Path = cookie_path, Expires = ngx.http_time(ngx.time() + session_expires_in)})
   ngx.header["X-Edo-Auth-Ta-Token"] = pre_sess:get_token()

   return respond_json({status = ngx.HTTP_UNAUTHORIZED, message = "start new session"})
end


-- リクエストを通す。
local function through(sess)
   -- リクエスト元 ID を添えて通す。

   local address = get_request_address()
   if not address then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "cannot get authenticated request source"})
   end

   -- リクエスト元のアドレスを特定した。
   ngx.log(log_level, "authenticated request source is " .. address)

   if address ~= sess:get_address() then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "authenticated request source is not session client"})
   end

   -- リクエスト元はセッション相手で間違いなかった。
   ngx.log(log_level, "authenticated request source is session client")

   local err = filter_cookie("X-Edo-Auth-Ta-Session", sess:get_id())
   if err then
      return respond_json(err)
   end
   ngx.req.set_header("X-Edo-Ta-Id", sess:get_ta())
end


-- 認証して通す。
local function authenticate(pre_sess, sess_db)
   local address = get_request_address()

   -- リクエスト元のアドレスを特定した。
   ngx.log(log_level, "authenticating request source is " .. address)
   if not address then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "cannot get authenticating request source"})
   end

   if address ~= pre_sess:get_address() then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "authenticating request source is not session client"})
   end

   -- リクエスト元はセッション相手で間違いなかった。
   ngx.log(log_level, "authenticating request source is session client")

   local ta_id = ngx.var.http_x_edo_auth_ta_id
   if not ta_id then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "no ta id"})
   end
   local token_sign = ngx.var.http_x_edo_auth_ta_token_sign
   if not token_sign then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "no token sign"})
   end
   local hash = ngx.var.http_x_edo_auth_hash_function or default_hash

   -- 認証情報が揃ってた。
   ngx.log(log_level, "authentication data was found")

   local public_key_pem, err = key_db.new_redis(sess_db.redis, redis_key_tag, key_directory, cache_expires_in):get(ta_id)
   if err then
      return respond_json(err)
   elseif not public_key_pem then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "no public key of " .. ta_id})
   end

   -- 公開鍵を取得できた。
   ngx.log(log_level, "public key of " .. ta_id .. " was found")

   local ok, err = verify(pre_sess:get_token(), token_sign, public_key_pem, hash)
   if err then
      return respond_json(err)
   elseif not ok then
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "verification failed"})
   end

   -- 認証できた。
   ngx.log(log_level, "authentication finished")

   local sess = session.new(pre_sess:get_id(), ta_id, pre_sess:get_address())
   local err = sess_db:save(sess, session_expires_in)
   if err then
      return respond_json(err)
   end

   -- セッションを認証済みに移行した。
   ngx.log(log_level, "session became authenticated")

   local err = filter_cookie("X-Edo-Auth-Ta-Session", sess:get_id())
   if err then
      return respond_json(err)
   end
   ngx.req.clear_header("X-Edo-Auth-Ta-Id")
   ngx.req.clear_header("X-Edo-Auth-Ta-Token-Sign")
   ngx.req.clear_header("X-Edo-Auth-Hash-Function")
   ngx.req.set_header("X-Edo-Ta-Id", ta_id)
end


-- ここから本編。


-- redis の準備。
local redis, err = redis_wrapper.new(redis_address, redis_timeout, redis_keepalive, redis_pool_size)
if err then
   return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
end


local session_id = ngx.var["cookie_x-edo-auth-ta-session"]
if not session_id then
   ngx.log(log_level, "session is not declared")
   local pre_sess_db = pre_session_db.new_redis(redis, redis_pre_session_tag)
   return start_session(pre_sess_db)
end

-- セッションが申告されてた。
ngx.log(log_level, "session is declared")

local sess_db = session_db.new_redis(redis, redis_session_tag)
local sess, err = sess_db:get(session_id)
if err then
   return respond_json(err)
elseif sess then
   -- 確立済みセッションがあった。
   ngx.log(log_level, "declared session is authenticated")
   return through(sess)
end

-- 確立済みセッションが無かった。
ngx.log(log_level, "declared session is not authenticated")

local pre_sess_db = pre_session_db.new_redis(redis, redis_pre_session_tag)
local pre_sess, err = pre_sess_db:get(session_id)
if err then
   return respond_json(err)
elseif not pre_sess then
   -- セッションが無かった。
   ngx.log(log_level, "declared session is not exist")
   return start_session(pre_sess_db)
end

-- セッションが確立されてなかった。
ngx.log(log_level, "declared session is authenticating")

return authenticate(pre_sess, sess_db)
