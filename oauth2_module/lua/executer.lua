return {
   execute = function(command)
      local file = io.popen(command, "r")
      local response = file:read("*a")
      file:close()
      return response
   end
}
