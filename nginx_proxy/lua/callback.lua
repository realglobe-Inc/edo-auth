package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local redis = require "redis"
local randomizer = require "randomizer"

local request_params = ngx.req.get_uri_args()
if request_params["state"] == ngx.var.cookie_oauth_state then
   request_params["client_id"] = ngx.var.client_id
   request_params["client_secret"] = ngx.var.client_secret
   request_params["redirect_uri"] = ngx.var.callback_url
   request_params["grant_type"] = "authorization_code"
   local request_body = ""
   for k, v in pairs(request_params) do
      if string.len(request_body) ~= 0 then
         request_body = request_body.."&"
      end
      request_body = request_body..k.."="..v
   end
   local options = {
      method = ngx.HTTP_POST,
      body = request_body
   }
   local local_proxy_namespace = ngx.var.local_proxy_namespace
   local access_token_endpoint = ngx.var.local_proxy_access_token_endpoint
   local res = ngx.location.capture(local_proxy_namespace..access_token_endpoint, options)

   require("cjson")
   local response_object = cjson.decode(res.body)

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
