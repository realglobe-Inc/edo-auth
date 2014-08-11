local json_safe = require "cjson.safe"
local logger = require "logger"

return {
   exit = function(message)
      logger.debug("exiter.lua", "exit:", message)
      ngx.header["X-EDO-Error-Message"] = message
      local response_object = {
         status = "error",
         error_message = message
      }
      ngx.status = 401
      ngx.header["Content-Type"] = "application/json"
      ngx.say(json_safe.encode(response_object))
      ngx.exit(ngx.status)
   end
}
