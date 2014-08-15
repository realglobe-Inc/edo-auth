package.path = package.path..";../lib/?.lua"

require "nginx_mock"
local cookie_manager = require "cookie_manager"

cookie_manager.set("foo", "bar", {})
assert(ngx.header["Set-Cookie"] == "foo=bar; secure; HttpOnly")

cookie_manager.set("hoge", "fuga", {path = "/"})
assert(ngx.header["Set-Cookie"] == "hoge=fuga; path=/; secure; HttpOnly")

cookie_manager.set("foobar", "hogefuga", {path = "/", expires = "xxxx"})
assert(ngx.header["Set-Cookie"] == "foobar=hogefuga; path=/; expires=xxxx; secure; HttpOnly")
