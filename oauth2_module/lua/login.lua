package.path = package.path..";"..ngx.var.lua_scripts_dir.."/?.lua"

local randomizer = require "randomizer"

local client_id = ngx.var.client_id
local oauth_authorize_url = ngx.var.authorize_url
local oauth_callback_url = ngx.var.callback_url
local scope = ngx.var.scope
local state = randomizer.generate(16)
local expires = ngx.cookie_time(tonumber(ngx.time()) + 300)
ngx.header["Set-Cookie"] = "oauth_state="..state..";path=/;expires="..expires

local request_params = {
   client_id = client_id,
   redirect_uri = oauth_callback_url,
   response_type = "code",
   scope = scope,
   state = state
}
local params_string = ""
for k, v in pairs(request_params) do
   if string.len(params_string) ~= 0 then
      params_string = params_string.."&"
   end
   params_string = params_string..k.."="..v
end
url = oauth_authorize_url.."?"..params_string
return ngx.redirect(url)
