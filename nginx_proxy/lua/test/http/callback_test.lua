package.path = package.path..";../?.lua;../lib/?.lua"

local config = require "config"

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.var.cookie_oauth_state = "hoge"
         require "callback"
      end
   )
)

assert(ngx.status == 401, "cookie の oauth_state が一致しないと 401")
package.loaded["nginx_mock"] = false
package.loaded["callback"] = false

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         require "callback"
      end
   )
)

assert(ngx.status == 302, "ステータスが "..ngx.status)
assert(ngx.header["Location"] == "/index.html", "リダイレクト先が "..ngx.header["Location"])
package.loaded["nginx_mock"] = false
package.loaded["callback"] = false
config.oauth.access_token_endpoint = "file://"..oauth_module_path.."/mock/error_response.json"

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         require "callback"
      end
   )
)

assert(ngx.status == 401, "アクセストークンの取得時にエラーが返ってきた場合は 401")
package.loaded["nginx_mock"] = false
package.loaded["callback"] = false
config.oauth.access_token_endpoint = "file://"..oauth_module_path.."/mock/invalid_response.json"

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         require "callback"
      end
   )
)

assert(ngx.status == 401, "アクセストークンが取得できなかった場合は 401")
