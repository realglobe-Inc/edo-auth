package.path = package.path..";../?.lua;../lib/?.lua"

require "nginx_mock"
require "login"

assert(ngx.status == 302, "ステータスが "..ngx.status)
assert(ngx.header["Location"] == "https://localhost/login/oauth/authorize?encoded_args")
