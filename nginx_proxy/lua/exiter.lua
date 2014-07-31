require "cjson"

return {
   exit = function(message)
      ngx.log(ngx.DEBUG, "exit: *** "..message.." ***")
      ngx.header["X-EDO-Error-Message"] = message
      local response_object = {
         status = "error",
         error_message = message
      }
      ngx.status = 401
      ngx.header["Content-Type"] = "application/json"
      ngx.say(cjson.encode(response_object))
      ngx.exit(ngx.status)
   end
}
