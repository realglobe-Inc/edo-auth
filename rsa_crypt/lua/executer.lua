return {
   execute = function(command)
      ngx.log(ngx.DEBUG, "execute: *** "..command.." ***")
      local file = io.popen(command, "r")
      local response = file:read("*a")
      file:close()
      ngx.log(ngx.DEBUG, "result: *** "..response.." ***")
      return response
   end
}
