package.path = package.path..";../lib/?.lua"

local redis = require "redis"
local client = redis:new()

client:clear()

assert(client:get("foo") == nil)
assert(client:setex("foo", "bar", 3600))
assert(client:get("foo") == "bar")
assert(client:setex("foo", "hoge", 3600))
assert(client:get("foo") == "hoge")
