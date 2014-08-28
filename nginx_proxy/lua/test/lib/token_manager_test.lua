package.path = package.path..";../lib/?.lua"

local redis = require "redis"
local token_manager = require "token_manager"
local client = redis:new()

client:clear()

assert(token_manager.access_token.get("foo") == nil)
assert(token_manager.access_token.set("foo", "bar", 3600))
assert(token_manager.access_token.get("foo") == "bar")
assert(token_manager.access_token.set("foo", "hoge", 3600))
assert(token_manager.access_token.get("foo") == "hoge")
assert(token_manager.id_token.get("foo") == nil)
assert(token_manager.id_token.set("foo", "bar", 3600))
assert(token_manager.id_token.get("foo") == "bar")
assert(token_manager.id_token.set("foo", "hoge", 3600))
assert(token_manager.id_token.get("foo") == "hoge")
