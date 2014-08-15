package.path = package.path..";../?.lua;../lib/?.lua"

require "nginx_mock"

ngx.header["X-OAUTH-ACCESS-TOKEN"] = "hoge"
require "add_header"
assert(ngx.header["X-OAUTH-ACCESS-TOKEN"] == nil, "X-OAUTH-ACCESS-TOKEN が nil になっていない")
package.loaded["add_header"] = nil


ngx.var.cookie_oauth_session_key = "foo"
require "add_header"
assert(ngx.header["X-OAUTH-ACCESS-TOKEN"], "X-OAUTH-ACCESS-TOKEN が nil になっている")
package.loaded["add_header"] = nil


ngx.var.cookie_oauth_session_key = "hoge"
require "add_header"
assert(ngx.header["X-OAUTH-ACCESS-TOKEN"] == nil, "X-OAUTH-ACCESS-TOKEN が nil になっていない")
package.loaded["add_header"] = nil
