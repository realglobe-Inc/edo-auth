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
local session = require("lib.auth_session")
local session_db = require("lib.auth_session_db")

-- $edo_log_level: デバッグログのレベル。
local log_level = varutil.get_level(ngx.var.edo_log_level)
-- $edo_redis_host: redis のアドレス。
local redis_host = ngx.var.edo_redis_host or "127.0.0.1"
-- $edo_redis_port: redis のポート。
local redis_port = ngx.var.edo_redis_port or 6379
-- $edo_redis_timeout: redis の接続待ち時間 (ミリ秒)。
local redis_timeout = ngx.var.edo_redis_timeout or 30 * 1000 -- 30 秒。
-- $edo_redis_keepalive: redis ソケットの待機時間 (ミリ秒)。
local redis_keepalive = ngx.var.edo_redis_keepalive or 60 * 1000 -- 1 分。
-- $edo_redis_pool_size: 1 ワーカー当たりの redis ソケット確保数。
-- 1 で十分かと思ったが、ab とかやってみるとそうではなさそう。
local redis_pool_size = ngx.var.edo_redis_pool_size or 16
-- $edo_session_tag: セッションを redis に格納する際のキーの接頭辞。
local redis_session_tag = ngx.var.edo_redis_session_tag or "auth.session"
-- $edo_session_label: セッションラベル。
local session_label = ngx.var.edo_session_label or "Auth-User"
-- $edo_backend_location: バックエンドに処理を渡すための location。
local backend_location = ngx.var.edo_backend_location or "@backend"


-- ここから本編。


local redis, err = redis_wrapper.new(redis_host, redis_port, redis_timeout, redis_keepalive, redis_pool_size)
if err then
   return erro.respond_html({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
end
local database = session_db.new_redis(redis, redis_session_tag)


local session_id = ngx.var["cookie_" .. session_label]
if not session_id then
   -- セッションが宣言されなかった。
   ngx.log(log_level, "no user session is declared")
   return ngx.exec(backend_location)
end

-- セッションが宣言された。
ngx.log(log_level, "user session is declared")

local session, err = database:get(session_id)
if err then
   return erro.respond_html({status = ngx.HTTP_INTERNAL_SERVER_ERROR, message = "database error: " .. err})
elseif not session then
   -- セッションが無かった。
   ngx.log(log_level, "declared session is not exist")
   return ngx.exec(backend_location)
end

-- セッションがあった。
ngx.log(log_level, "declared session is exist")

ngx.req.set_header("X-Auth-User", session:get_account())

return;
