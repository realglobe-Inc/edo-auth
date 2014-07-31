package.path = package.path..";"..ngx.var.lua_scripts_dir.."/?.lua"

local redis = require "redis"

ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", nil)

local session_key = ngx.var.cookie_oauth_session_key
if session_key and session_key ~= "" then
   local value = redis.get(session_key)
   if value and value ~= "" then
      ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", value)
   end
end
