-- ユーザー認証済みなら、ヘッダを設定して通す。
-- ユーザー認証の途中なら、ユーザー認証を終わらせる。
--   認証できたら、ユーザー認証済みセッションを開始して同じページにリダイレクトさせる。
--   認証できなかったら、403 Forbidden とか。
-- サービス認証情報があったら、サービス認証を試みる。
--   認証できたら、ヘッダを設定して通す。
--   認証できなかったら、403 Forbidden とか。
-- 以上に該当しなければ、ユーザー認証を開始させる。

-- ユーザー認証で通った場合、ヘッダに X-OAUTH-ACCESS-TOKEN と X-OAUTH-ID-PROVIDER-UUID が付く。
-- X-EDO-Sender-UUID と X-EDO-Receiver-UUID があれば取り除かれる。
-- サービス認証で通った場合、X-EDO-Sender-UUID と X-EDO-Receiver-UUID が付く。
-- X-OAUTH-ACCESS-TOKEN と X-OAUTH-ID-PROVIDER-UUID はそのまま。

-- 内部的なエラーの形式は
-- { status = <HTTP のステータスコード>, message = <適当なメッセージ> }
-- の形で正常な返り値の後に返すようにする。
-- 関数内で status が決められないような場合は status の未定義を許す。

package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local config = require("config")
local exiter = require("exiter")
local logger = require("logger")
local public_key_manager = require("public_key_manager")
local user = require("user")
local session_manager = require("session_manager")
local id_provider = require("id_provider")
local service = require("service")

-- % エンコードをデコードする。
local function decode(str)
   if not str then
      return nil
   end
   str = string.gsub(str, "+", " ")
   str = string.gsub(str, "%%(%x%x)", function(h) return string.char(tonumber(h,16)) end)
   return str
end

local session_key = ngx.var.cookie_oauth_session_key
if session_key then
   -- セッションあり。
   logger.debug("oauth session key was found")

   local session, err = session_manager.get(session_key)
   if err then
      return exiter.exit_with_status(err.status, err.message)
   elseif session then
      -- ユーザー認証済み。
      logger.debug("authorized user")

      -- セッション情報をフィルタする。
      -- TODO へたくそ。何か上手い方法は無いか？
      local filtered_cookie = string.gsub(ngx.req.get_headers()["Cookie"], session_key, "")
      ngx.req.set_header("Cookie", filtered_cookie)

      logger.debug("session key was filtered")

      user.set_header(session)
      return
   end
end

local client_id = config.oauth.client_id
if not client_id then
   client_id = config.service_uuid
end

if not client_id then
   return exiter.exit_with_status(ngx.HTTP_INTERNAL_SERVER_ERROR, "no client id")
end

local code = decode(ngx.var.arg_code) -- TODO code は一般的過ぎるから、変な名前に変えた方が良さそう。
if code then

   -- ユーザー認証の途中。
   logger.debug("user authorization continues")

   local access_token_uri = nil

   local id_provider_uuid = decode(ngx.var.arg_id_provider_uuid)
   local id_provider_registry_uri = config.oauth.id_provider_registry_uri
   if id_provider_uuid and id_provider_registry_uri then
      access_token_uri, err = id_provider.get_access_token_uri(id_provider_registry_uri, id_provider_uuid)
      if err then
         return exiter.exit_with_status(err.status, err.message)
      elseif not access_token_uri then
         return exiter.exit_with_status(ngx.HTTP_FORBIDDEN, "id provider " .. id_provider_uuid .. " is not exist")
      end

      -- アクセストークン取得用 URI が分かった。
      logger.debug("access token uri is " .. access_token_uri)
   end

   if not access_token_uri then
      access_token_uri = config.oauth.access_token_endpoint
   end

   if not access_token_uri then
      -- code があるのに発行者が指定されていないときはクライアントが悪い。
      return exiter.exit_with_status(ngx.HTTP_BAD_REQUEST, "no id_provider_uuid parameter")
   end

   local private_key_uuid = config.service_uuid
   local private_key_directory = config.oauth.private_key_directory
   local client_secret = config.oauth.client_secret

   local access_token, err = id_provider.get_access_token(access_token_uri, client_id, code, private_key_uuid, private_key_directory, client_secret)
   if err then
      return exiter.exit_with_status(err.status, err.message)
   end

   -- アクセストークン貰えた。
   logger.debug("access token was published from " .. access_token_uri)

   local session, err = session_manager.generate_session(access_token, nil, id_provider_uuid)
   if err then
      return exiter.exit_with_status(err.status, err.message)
   end

   -- セッションを生成した。
   logger.debug("session was generated")

   return user.start_session(session)
end

local request_headers = ngx.req.get_headers()
local private_key_uuid = request_headers["X-EDO-Private-Key-UUID"]
if private_key_uuid then
   -- サービス認証する。
   logger.debug("try service authorization")

   local service_uuid = config.service_uuid
   if not service_uuid then
      return exiter.exit_with_status(ngx.HTTP_INTERNAL_SERVER_ERROR, "no service uuid")
   end
   local token_lifetime = nil
   if config.rsa.token_lifetime then
      token_lifetime = tonumber(config.rsa.token_lifetime)
   end

   local public_key_string = public_key_manager.get(private_key_uuid) -- 実行コスト的には authorize の直前の方が良い。
   if not public_key_string then
      return exiter.exit_with_status(ngx.HTTP_FORBIDDEN, "no public key " .. private_key_uuid)
   end

   local encoded_token = request_headers["X-EDO-Auth-Encoded-Token"]
   if encoded_token then
      -- 直接暗号方式。
      local auth_info, err = service.authorize(public_key_string, encoded_token, service_uuid, token_lifetime)
      if err then
         return exiter.exit_with_status(err.status, err.message)
      end

      -- 認証できた。
      logger.debug(auth_info.sender_uuid .. " was authorized with direct sign")

      service.set_header(auth_info)
      return
   end

   local raw_token = request_headers["X-EDO-Auth-Token"]
   local signed_token = request_headers["X-EDO-Auth-Signed-Token"]
   local hash_function = request_headers["X-EDO-Hash-Function"]
   if raw_token and signed_token then
      -- ハッシュ値方式。
      local auth_info, err = service.authorize_sign(public_key_string, raw_token, signed_token, hash_function, service_uuid, token_lifetime)
      if err then
         return exiter.exit_with_status(err.status, err.message)
      end

      -- 認証できた。
      logger.debug(auth_info.sender_uuid .. " was authorized with hashed sign")

      service.set_header_sign(auth_info)
      return
   end

   -- X-EDO-Private-Key-UUID があるのに他が揃っていないときはクライアントが悪い。
   return exiter.exit_with_status(ngx.HTTP_BAD_REQUEST, "lack of service authorization headers")
end

-- ユーザー認証が必要。
logger.debug("begin user authorization")

local authorize_uri = config.oauth.authorize_url
if not authorize_uri then
   return exiter.exit_with_status(ngx.HTTP_INTERNAL_SERVER_ERROR, "no user authorization uri")
end

return user.authorize(client_id, authorize_uri)
