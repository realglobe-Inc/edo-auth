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
   end,

   exit_with_status = function(status, message)
      logger.notice("exiter.exit_with_status:", status, "," .. message)

      status = status or ngx.HTTP_INTERNAL_SERVER_ERROR

      ngx.header["X-EDO-Error-Message"] = message
      local response_object = {
         status = "error",
         status_code = status,
         error_message = message
      }
      ngx.status = status
      ngx.header["Content-Type"] = "application/json"
      local body = json_safe.encode(response_object)
      ngx.header["Content-Length"] = string.len(body) .. ""
      ngx.print(body)
      return ngx.exit(status)
   end
}
