package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local logger = require "logger"
local redis = require "redis"
local randomizer = require "randomizer"
local curl_wrapper = require "curl_wrapper"

local request_params = ngx.req.get_uri_args()
if request_params["state"] == ngx.var.cookie_oauth_state then
   -- TODO
   local access_token_endpoint = "https://github.com/login/oauth/access_token"
   local req_params = {
      client_id = ngx.var.client_id,
      client_secret = ngx.var.client_secret,
      code = request_params["code"]
   }
   local request_body = ngx.encode_args(req_params)
   local response = curl_wrapper.post(access_token_endpoint, request_body)
   logger.debug("callback.lua", "response.body:", response.body)
   local response_object = ngx.decode_args(response.body)

   local session_key = randomizer.generate(32)
   local access_token = response_object["access_token"]
   local expires_in = response_object["expires_in"]
   if expires_in == nil then
      expires_in = ngx.var.default_expires_in
   end
   local expires = ngx.cookie_time(tonumber(ngx.time()) + expires_in)
   if redis.setex(session_key, access_token, expires_in) then
      ngx.header["Set-Cookie"] = "oauth_session_key="..session_key..";path=/;expires="..expires
      ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", access_token)
   else
      -- TODO
   end
else
   -- TODO
end
return ngx.redirect("/")
