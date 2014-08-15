package.path = package.path..";../lib/?.lua"

require "nginx_mock"
local exiter = require "exiter"

exit_code = coroutine.resume(
   coroutine.create(
      function()
         exiter.exit("test")
      end
   )
)
assert(ngx.status == 401, "ステータスが 401 ではなく "..ngx.status)
assert(exit_code == false, "exit_code が false")
