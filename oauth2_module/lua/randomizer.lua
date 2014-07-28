local executer = require "executer"

return {
   generate = function(length)
      return executer.execute('head /dev/urandom | tr -c -d "[:alnum:]" | head -c '..length)
   end
}
