package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local config = require "config"
local logger = require "logger"

return {
   set = function(name, value, option_params)
      local cookie_string = name.."="..value
      for key, val in pairs(option_params) do
         cookie_string = cookie_string.."; "..key.."="..val
      end
      if config.cookie.secure then
         cookie_string = cookie_string.."; secure"
      end
      if config.cookie.http_only then
         cookie_string = cookie_string.."; HttpOnly"
      end
      logger.debug("cookie_manager.lua", "cookie_string =", cookie_string)
      ngx.header["Set-Cookie"] = cookie_string
   end
}
