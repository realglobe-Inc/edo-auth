package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local json_safe = require "cjson.safe"
local config = require "config"
local logger = require "logger"
local redis = require "redis"
local openssl_rsa = require "openssl_rsa"
local curl_wrapper = require "curl_wrapper"
local cookie_manager = require "cookie_manager"
local exiter = require "exiter"

local request_params = ngx.req.get_uri_args()
if request_params["state"] == ngx.var.cookie_oauth_state then
   local req_params = {
      client_id = config.oauth.client_id,
      client_secret = config.oauth.client_secret,
      grant_type = "authorization_code",
      redirect_url = config.oauth.callback_url,
      code = request_params["code"]
   }
   local request_body = ngx.encode_args(req_params)
   curl_wrapper.set_headers({"Accept: application/json"})
   local response = curl_wrapper.post(config.oauth.access_token_endpoint, request_body)
   logger.debug("callback.lua", "response.body:", response.body)
   local response_object = json_safe.decode(response.body)

   local session_key = openssl_rsa.random_string(32)
   local access_token = response_object["access_token"]
   local expires_in = response_object["expires_in"]
   if not expires_in then
      expires_in = config.oauth.access_token_default_expires_in
   end
   local expires = ngx.cookie_time(tonumber(os.date("%s")) + expires_in)
   logger.debug("callback.lua", "redis", redis)
   local client = redis:new()
   logger.debug("callback.lua", "client", client)
   if client:setex(session_key, access_token, expires_in) then
      cookie_manager.set("oauth_session_key", session_key, {path = "/", expires = expires})
      ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", access_token)
   else
      -- TODO
   end
else
   exiter.exit("invalid state")
end

ngx.redirect(config.oauth.after_login_path)
