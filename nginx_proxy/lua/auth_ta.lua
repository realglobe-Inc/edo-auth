-- クライアント認証型 TA 認証だけのプロキシ。

local openssl = require("openssl")
local cjson = require("cjson.safe")
local redis = require("resty.redis")


local function get_log_level(level)
   if (not level) or level == "debug" then
      return ngx.DEBUG
   elseif level == "info" then
      return ngx.INFO
   elseif level == "notice" then
      return ngx.NOTICE
   elseif level == "warn" then
      return ngx.WARN
   elseif level == "error" then
      return ngx.ERR
   elseif level == "crit" then
      return ngx.CRIT
   elseif level == "alert" then
      return ngx.ALERT
   elseif level == "emerg" then
      return ngx.EMERG
   else
      return ngx.DEBUG
   end
end

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


-- $edo_auth_log_level: デバッグログのレベル。
local log_level = get_log_level(ngx.var.edo_auth_log_level)
-- $edo_auth_session_length: セッション ID の文字列長。
local session_length = ngx.var.edo_auth_session_length or 40
-- $edo_auth_token_length: 署名用文字列の長さ。
local token_length = ngx.var.edo_auth_token_length or 40
-- $edo_auth_default_hash: 無指定時の署名検証に使われるハッシュ関数。
local default_hash = ngx.var.edo_auth_default_hash or "sha256"
-- $edo_auth_start_delay: セッションの開始から認証完了までの猶予期間。
local start_delay = ngx.var.edo_auth_start_delay or 3 * 60 -- 3 分。
-- $edo_auth_redis_host: redis のアドレス。
local redis_host = ngx.var.edo_auth_redis_host or "127.0.0.1"
-- $edo_auth_redis_port: redis のポート。
local redis_port = ngx.var.edo_auth_redis_port or 6379
-- $edo_auth_redis_connection_keepalive: redis ソケットの待機時間 (ミリ秒)。
local redis_connection_keepalive = ngx.var.edo_auth_redis_connection_keepalive or 60 * 1000 -- 1 分。
-- $edo_auth_redis_connection_pool: 1 ワーカー当たりの redis ソケット確保数。
-- 1 で十分かと思ったが、ab とかやってみるとそうではなさそう。
local redis_connection_pool = ngx.var.edo_auth_redis_connection_pool or 16
-- $edo_auth_public_key_directory: 公開鍵が <TA の ID>.pub.pem って名前で入ってるディレクトリ。
local public_key_directory = get_absolute_path(ngx.var.edo_auth_public_key_directory)
-- $edo_auth_session_expires_in: セッションの有効期間。
local session_expires_in = ngx.var.edo_auth_session_expires_in or 60 * 60 -- 1 時間。
-- $edo_auth_cache_expires_in: キャッシュの有効期間。
local cache_expires_in = ngx.var.edo_auth_cache_expires_in or 10 * 60 -- 10 分。
-- $edo_auth_cookie_path: cookie の Path。複数 TA を混在させるならあった方が良いか？
local cookie_path = ngx.var.edo_auth_cookie_path or "/"


-- エラーレスポンスを返す。
local function exit(err, redis_client)
   if redis_client then
      local ok, err = redis_client:set_keepalive(redis_connection_keepalive, redis_connection_pool)
      if not ok then
         ngx.log(ngx.ERR, err)
      end
   end

   err.status = err.status or ngx.HTTP_INTERNAL_SERVER_ERROR

   local body, e = cjson.encode(err)
   if e then
      ngx.log(ngx.ERR, e.message)
      body = "{status=" .. err.status .. ",message=\"json encoding error\"}"
   end

   -- レスポンスの中身を準備できた。
   ngx.log(ngx.ERR, "respond " .. err.status .. ": " .. err.message)

   ngx.header["X-Edo-Ta-Auth-Error"] = err.message
   ngx.header["Content-Type"] = "application/json"
   ngx.header["Content-Length"] = string.len(body) .. ""
   ngx.status = err.status
   ngx.print(body)
   return ngx.exit(err.status)
end


-- ランダムな文字列を生成する。
local function generate_random_string(length)
   -- OpenSSL の疑似乱数列を BASE64 エンコードする。
   local buff, err = openssl.random((length * 3  + 3) / 4, false)
   if not err then -- よく分からないが、成功すると true ？
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
   -- TODO へたくそ。何か上手い方法は無いか？
   local after = string.gsub(ngx.var.http_cookie, value, "")
   ngx.req.set_header("Cookie", after)
end


-- 署名検証。
local function verify(token, sign, public_key_pem, hash)
   local raw_sign = ngx.decode_base64(sign)
   if not raw_sign then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "bad sign format"}
   end

   local public_key, err = openssl.pkey.read(public_key_pem, true)
   if err then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = err}
   end
   return public_key:verify(token, raw_sign, hash)
end


-- redis の準備。
local redis_client, err = redis.new()
if err then
   return exit({message = err})
end


local session_manager = {
   -- 未認証のセッションを登録する。
   -- {
   --     id      = <セッション ID>,
   --     token   = <署名用乱数列>,
   --     client  = <通信相手のアドレス>,
   -- }
   add_unauthenticated = function(session, expires_in)
      local buff, err = cjson.encode(session)
      if err then
         return {message = err}
      end
      local _, err = redis_client:setex("session:unauthenticated:" .. session.id, expires_in, buff)
      if err then
         return {message = err}
      end
   end,

   -- 認証済みのセッションを登録する。
   -- {
   --     auth    = true,
   --     id      = <セッション ID>,
   --     ta      = <通信先 TA の ID>
   --     client  = <通信先のアドレス>,
   -- }
   add_authenticated = function(session, expires_in)
      local buff, err = cjson.encode(session)
      if err then
         return {message = err}
      end
      local _, err = redis_client:setex("session:authenticated:" .. session.id, expires_in, buff)
      if err then
         return {message = err}
      end
   end,

   -- 未認証のセッションなら消して返す。
   -- 認証済みセッションなら消さずに返す。
   get = function(session_id)
      local buff, err = redis_client:get("session:authenticated:" .. session_id)
      if err then
         return nil, {message = err}
      elseif buff and buff ~= ngx.null then
         ngx.log(log_level, "authenticated session is exist")

         local session, err = cjson.decode(buff)
         if err then
            return nil, {message = err}
         end
         -- TODO セッションの期限を更新すべきか？更新したら Set-Cookie でクライアントに報せなきゃならない。
         return session
      end

      -- 認証済みセッションではなかった。

      -- redis に取り出しと削除をアトミックにやるようなコマンドが無さそうだったので、代わりにダミーに置き換える。
      local buff, err = redis_client:getset("session:unauthenticated:" .. session_id, "")
      if err then
         return nil, {message = err}
      elseif (not buff) or buff == ngx.null or buff == "" then
         -- そんなセッションは無かった。
         ngx.log(log_level, "session is not exist")
         return nil
      end

      -- 取り出せた。
      ngx.log(log_level, "authenticating session is exist")

      -- ダミーを消す。
      local _, err = redis_client:del("session:unauthenticated:" .. session_id)
      if err then
         ngx.log(ngx.ERR, err)
      end

      local session, err = cjson.decode(buff)
      if err then
         return nil, {message = err}
      end
      return session
   end
}


local public_key_manager = {
   -- 公開鍵を取得する。
   get = function(ta_id)
      -- redis を調べてから、ファイルを調べる。

      local public_key_pem = redis_client:get("public_key:" .. ta_id)
      if public_key_pem and public_key_pem ~= ngx.null then
         ngx.log(log_level, "public_key is exist in redis")
         -- 公開鍵の更新時に再起動が必須になってしまうので、キャッシュの期限は更新しない方が良さそう。
         return public_key_pem
      end

      -- redis になかった。

      local public_key_path = public_key_directory .. "/" .. ta_id .. ".pub.pem"
      local fd, err = io.open(public_key_path)
      if err then
         -- 公開鍵ファイルが無かった？
         ngx.log(ngx.ERR, err)
         return nil
      end

      -- 公開鍵ファイルがあった。
      ngx.log(log_level, "public key file " .. ta_id .. ".pub.pem is exist")

      public_key_pem = fd:read("*a")
      fd:close()
      if (not public_key_pem) or public_key_pem == "" then
         return nil
      end

      -- 公開鍵が読めた。
      ngx.log(log_level, "public key is exist in " .. ta_id .. ".pub.pem")

      -- キャッシュする。
      local _, err = redis_client:setex("public_key:" .. ta_id, cache_expires_in, public_key_pem)
      if err then
         ngx.log(ngx.ERR, err)
      else
         ngx.log(log_level, "public key was cached into redis")
      end

      return public_key_pem
   end
}


-- セッションを開始する。
local function start_session()
   -- 乱数列、セッション ID を生成し、
   -- セッション ID にリクエスト元アドレスと乱数列を紐付けて保存し、
   -- 401 Unauthorized でセッション ID と乱数列を返す。

   local client = get_request_address()
   if not client then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "cannot get unauthenticated request source"}, redis_client)
   end

   -- リクエスト元のアドレスを特定した。
   ngx.log(log_level, "unauthenticated request source is " .. client)

   local session_id, err = generate_random_string(session_length)
   if err then
      return exit(err, redis_client)
   end

   -- セッション ID を生成した。
   ngx.log(log_level, "session id was generated")

   local token, err = generate_random_string(token_length)
   if err then
      return exit(err, redis_client)
   end

   -- 署名用乱数列を生成した。
   ngx.log(log_level, "token was generated")

   local session = {
      id = session_id,
      token = token,
      client = client,
   }

   local err = session_manager.add_unauthenticated(session, start_delay)
   if err then
      return exit(err, redis_client)
   end

   -- セッションを登録した。
   ngx.log(log_level, "session began")

   set_cookie("X-Edo-Ta-Session", session.id, {Path = cookie_path, Expires = ngx.http_time(ngx.time() + session_expires_in)})
   ngx.header["X-Edo-Ta-Token"] = session.token

   return exit({status = ngx.HTTP_UNAUTHORIZED, message = "start new session"}, redis_client)
end


-- リクエストを通す。
local function through(session)
   -- リクエスト元 ID を添えて通す。

   local client = get_request_address()
   if not client then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "cannot get authenticated request source"}, redis_client)
   end

   -- リクエスト元のアドレスを特定した。
   ngx.log(log_level, "authenticated request source is " .. client)

   if client ~= session.client then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "authenticated request source is not session client"}, redis_client)
   end

   -- リクエスト元はセッション相手で間違いなかった。
   ngx.log(log_level, "authenticated request source is session client")

   filter_cookie("X-Edo-Ta-Session", session.id)
   ngx.req.set_header("X-Edo-Ta-Id", session.ta)

   local ok, err = redis_client:set_keepalive(redis_connection_keepalive, redis_connection_pool)
   if not ok then
      ngx.log(ngx.ERR, err)
   end
end


-- 認証して通す。
local function authenticate(session)
   local client = get_request_address()

   -- リクエスト元のアドレスを特定した。
   ngx.log(log_level, "authenticating request source is " .. client)
   if not client then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "cannot get authenticating request source"}, redis_client)
   end

   if client ~= session.client then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "authenticating request source is not session client"}, redis_client)
   end

   -- リクエスト元はセッション相手で間違いなかった。
   ngx.log(log_level, "authenticating request source is session client")

   local ta_id = ngx.var.http_x_edo_ta_id
   if not ta_id then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "no ta id"}, redis_client)
   end
   local token_sign = ngx.var.http_x_edo_ta_token_sign
   if not token_sign then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "no token sign"}, redis_client)
   end
   local hash = ngx.var.http_x_edo_hash_function or default_hash

   -- 認証情報が揃ってた。
   ngx.log(log_level, "authentication data was found")

   local public_key_pem, err = public_key_manager.get(ta_id)
   if err then
      return exit(err, redis_client)
   elseif not public_key_pem then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "no public key of " .. ta_id}, redis_client)
   end

   -- 公開鍵を取得できた。
   ngx.log(log_level, "public key of " .. ta_id .. " was found")

   local ok, err = verify(session.token, token_sign, public_key_pem, hash)
   if err then
      return exit(err, redis_client)
   elseif not ok then
      return exit({status = ngx.HTTP_FORBIDDEN, message = "verification failed"}, redis_client)
   end

   -- 認証できた。
   ngx.log(log_level, "authentication finished")

   session.token = nil
   session.auth = true
   session.ta = ta_id
   local err = session_manager.add_authenticated(session, session_expires_in)
   if err then
      return exit(err, redis_client)
   end

   -- セッションを認証済みに移行した。
   ngx.log(log_level, "session became authenticated")

   filter_cookie("X-Edo-Ta-Session", session.id)
   ngx.req.clear_header("X-Edo-Ta-Token-Sign")
   ngx.req.clear_header("X-Edo-Hash-Function")

   local ok, err = redis_client:set_keepalive(redis_connection_keepalive, redis_connection_pool)
   if not ok then
      ngx.log(ngx.ERR, err)
   end
end


-- ここから本編。


-- finally とか defer が無いので、
-- exit, start_session, through, authenticate で close する。
local ok, err = redis_client:connect(redis_host, redis_port)
if not ok then
   return exit({message = err})
end


local session_id = ngx.var["cookie_x-edo-ta-session"]
if not session_id then
   ngx.log(log_level, "session is not declared")
   return start_session()
end

-- セッションが申告されてた。
ngx.log(log_level, "session is declared")

local session, err = session_manager.get(session_id)
if err then
   return exit(err, redis_client)
elseif not session then
   -- セッションが無かった。
   ngx.log(log_level, "declared session was not found")
   return start_session()
end

-- セッションがあった。
ngx.log(log_level, "declared session was found")

if session.auth then
   return through(session)
end

-- まだ認証されてなかった。
ngx.log(log_level, "declared session is authenticating")

return authenticate(session)