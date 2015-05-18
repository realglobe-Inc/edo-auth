-- Copyright 2015 realglobe, Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

local test = require("test.test")
local redis_wrapper = require("lib.redis_wrapper")


-- 成功したら 200 OK を返す。

local client, err = redis_wrapper.new(ngx.var.redis_host, ngx.var.redis_port, 1000, 10 * 1000, 16)
if err then
   return test.response_error("new failed: " .. err)
end

local test = function(k, v)
   local _, err = client:connect()
   if err then
      return test.response_error("connect failed: " .. err)
   end

   local _, err = client.base:set(k, v, "ex", 1)
   if err then
      return test.response_error("set failed: " .. err)
   end

   local buff, err = client.base:get(k)
   if err then
      return test.response_error("get failed: " .. err)
   elseif not buff then
      return test.response_error("get failed: no value of " .. k)
   elseif buff ~= v then
      return test.response_error("get failed: value of " .. k .. " is " .. buff .. " not " .. v)
   end

   local _, err = client:close()
   if err then
      return test.response_error("close failed: " .. err)
   end
end

test("a", "b")
test("a", "c")

ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
