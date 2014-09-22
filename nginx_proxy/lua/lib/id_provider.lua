-- ID プロバイダ関係。

local private_key_manager = require("private_key_manager")
local json_safe = require("cjson.safe")
local openssl_rsa = require("openssl_rsa")
local logger = require("logger")
local curl_wrapper = require("curl_wrapper")

-- ID プロバイダのアクセストークン URI を取得する。
local function get_access_token_uri(id_provider_registry_uri, id_provider_uuid)
   local curl_client = curl_wrapper:new()
   local uri = id_provider_registry_uri .. id_provider_uuid .. "/access_token_uri"
   local response = curl_client:get(uri)

   if response.code ~= ngx.HTTP_OK then
      return nil, {status = response.code, message = "cannot get access token uri"}
   end

   logger.debug("id_provider.lua", "response for access token uri is ", response.body)

   local response_object = json_safe.decode(response.body)
   if not response_object.id_provider then
      return nil, {message = "cannot parse id_provider"}
   elseif not response_object.id_provider.access_token_uri then
      return nil, {message = "cannot parse id_provider.access_token_uri"}
   else
      return response_object.id_provider.access_token_uri
   end
end

-- ID プロバイダにアクセストークンを発行してもらう。
-- access_token
-- {
--   id = <アクセストークン本体>,
--   expires_in = <有効期間 (期限じゃない)>
-- }
local function get_access_token(access_token_uri, client_id, code, private_key_uuid, private_key_directory, client_secret)

   if not client_secret then
      -- code に署名して client_secret とする。

      local private_key_string, err = private_key_manager.get(private_key_uuid, private_key_directory)
      if err then
         return nil, err
      end

      -- 秘密鍵はあった。
      logger.debug("id_provider.lua", "private key was read")

      client_secret = openssl_rsa.sign_base64(code, private_key_string)
      if not client_secret then
         return nil, { message = "signing code failed" }
      end
   end

   -- client_secret が準備できた。
   logger.debug("id_provider.lua", "client_secret is OK")

   local curl_client = curl_wrapper:new()
   curl_client:set_headers({"Accept: application/json"})

   local redirect_uri = ngx.var.scheme .. "://" .. ngx.var.http_host .. ngx.var.request_uri
   local access_token_request_params = {
      client_id = client_id,
      client_secret = client_secret,
      redirect_uri = redirect_uri,
      code = code
   }
   local request_body = ngx.encode_args(access_token_request_params)

   local response = curl_client:post(access_token_uri, request_body)

   if response.code ~= ngx.HTTP_OK then
      return nil, {status = response.code, message = "cannot get access token"}
   end

   logger.debug("id_provider.lua", "response for access token is", response.body)

   local response_object = json_safe.decode(response.body)
   if response_object.error then
      local error_message = response_object.error
      if response_object.error_description then
         error_message = error_message .. ": " .. response_object.error_description
      end
      return nil, {status = ngx.HTTP_FORBIDDEN, message = error_message}
   end

   if not response_object.access_token then
      return nil, {message = "no access token"}
   end

   -- access token を取得できた。
   logger.debug("id_provider.lua", "access token is OK")

   local access_token = {id = response_object.access_token}
   if response_object.expires_in then
      access_token.expires_in = response_object.expires_in
   end

   return access_token
end

return {
   get_access_token_uri = get_access_token_uri,
   get_access_token = get_access_token
}
