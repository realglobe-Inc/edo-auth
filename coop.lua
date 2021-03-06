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


-- ユーザー認証代行。

local varutil = require("lib.varutil")
local erro = require("lib.erro")
local redis_wrapper = require("lib.redis_wrapper")
local session = require("lib.coop_session")
local session_db = require("lib.coop_session_db")


-- $edo_log_level: デバッグログのレベル。
local log_level = varutil.get_level(ngx.var.edo_log_level)
-- $edo_redis_address: redis のアドレス。
local redis_address = ngx.var.edo_redis_address or "127.0.0.1:6379"
-- $edo_redis_timeout: redis の接続待ち時間 (ミリ秒)。
local redis_timeout = ngx.var.edo_redis_timeout or 30 * 1000 -- 30 秒。
-- $edo_redis_keepalive: redis ソケットの待機時間 (ミリ秒)。
local redis_keepalive = ngx.var.edo_redis_keepalive or 60 * 1000 -- 1 分。
-- $edo_redis_pool_size: 1 ワーカー当たりの待機させる redis 接続数。
local redis_pool_size = ngx.var.edo_redis_pool_size or 16
-- $edo_redis_session_tag: セッションを redis に格納する際のキーの接頭辞。
local redis_session_tag = ngx.var.edo_redis_session_tag or "csession"
-- $edo_backend_location: バックエンドに処理を渡すための location。
local backend_location = ngx.var.edo_coop_location or "/api_backend"


-- Set-Cookie ヘッダの内容からセッション ID と有効期限を取得する。
local function get_session(cookie)
   if not cookie then
      return;
   end
   local id, exp_in
   for k, v in cookie:gmatch(" ?([^;]+)=([^;]+)") do
      if k == "Edo-Cooperation" then
         id = v
      elseif k == "Expires" then
         exp_in = ngx.parse_http_time(v) - ngx.time()
      elseif k == "MaxAge" then
         exp_in = v
      end
   end
   return id, exp_in
end


-- X-Edo-Cooperation-Error ヘッダを付けて返信する。
local function respond_json(params)
   if params.message then
      ngx.header["X-Edo-Cooperation-Error"] = params.message
   else
      ngx.header["X-Edo-Cooperation-Error"] = "error occurred"
   end
   return erro.respond_json(params)
end


-- ここから本編。


local session_id = ngx.var["cookie_edo-cooperation"]
if session_id then
   -- セッションが宣言された。
   ngx.log(log_level, "TA session is declared")

   local redis, err = redis_wrapper.new(redis_address, redis_timeout, redis_keepalive, redis_pool_size)
   if err then
      return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
   end
   local database = session_db.new_redis(redis, redis_session_tag)

   local session, err = database:get(session_id)
   if err then
      return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
   elseif not session then
      -- セッションが無かった。
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "invalid session"})
   end

   -- セッションがあった。
   ngx.log(log_level, "declared session is exist")

   ngx.req.set_header("X-Auth-User", session:get_account())
   ngx.req.set_header("X-Auth-User-Tag", session:get_account_tag())
   ngx.req.set_header("X-Auth-From-Id", session:get_from_ta())
   local accounts = session:get_accounts()
   if accounts then
      ngx.req.set_header("X-Auth-Users", accounts)
   end
   return;
end


-- セッションが宣言されなかった。

local params
if not ngx.var.http_x_edo_code_tokens then
   for key, val in pairs(ngx.req.get_uri_args()) do
      if key == "code_tokens" then
         params = {
            ["code_tokens"] = val
         }
         break
      end
   end

   if not params then
      -- 仲介コードが無かった。
      return respond_json({status = ngx.HTTP_FORBIDDEN, message = "no code tokens"})
   end
end

-- バックエンドに処理を投げる。
local resp = ngx.location.capture(backend_location, params)

local account_info = resp.header["X-Auth-User"]
if not account_info then
   -- 取得失敗。
   -- ステータスのコピー。
   ngx.status = resp.status

   -- ヘッダのコピー。
   for k, _ in pairs(ngx.header) do
      ngx.header[k] = nil
   end
   for k, v in pairs(resp.header) do
      ngx.header[k] = v
   end

   -- 中身のコピー。
   if resp.body then
      ngx.say(resp.body)
   end

   ngx.exit(ngx.status)
   return;
end

-- 取得できた。
ngx.log(log_level, "got account info")

if not params then
   ngx.req.clear_header("X-Edo-Code-Tokens")
end

ngx.req.set_header("X-Auth-User", account_info)
local account_tag = resp.header["X-Auth-User-Tag"]
if not account_tag then
   return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "no account tag"})
end
ngx.req.set_header("X-Auth-User-Tag", account_tag)
local from_ta = resp.header["X-Auth-From-Id"]
if not from_ta then
   return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "no from-TA"})
end
ngx.req.set_header("X-Auth-From-Id", from_ta)
local accounts_info = resp.header["X-Auth-Users"]
if accounts_info then
   ngx.req.set_header("X-Auth-Users", accounts_info)
end

local session_cookie = resp.header["Set-Cookie"]
local session_id, session_exp_in = get_session(session_cookie)
if session_id and session_exp_in and session_exp_in > 0 then
   -- セッションが宣言された。
   ngx.log(log_level, "TA session is declared")

   local redis, err = redis_wrapper.new(redis_address, redis_timeout, redis_keepalive, redis_pool_size)
   if err then
      return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
   end
   local database = session_db.new_redis(redis, redis_session_tag)

   local err = database:save(session.new(session_id, account_info, account_tag, from_ta, accounts_info), session_exp_in)
   if err then
      return respond_json({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
   end

   ngx.log(log_level, "saved account info")

   ngx.header["Set-Cookie"] = session_cookie
end
