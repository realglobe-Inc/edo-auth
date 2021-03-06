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

local redis_wrapper = require("lib.redis_wrapper")
local test = require("test.test")
local tutil = require("lib.table")
local session = require("lib.auth_session")
local session_db = require("lib.auth_session_db")


-- redis_address: redis のアドレス。
-- 成功したら 200 OK を返す。

local client, err = redis_wrapper.new(ngx.var.redis_address, 1000, 10 * 1000, 16)
if err then
   return test.response_error("redis_wrapper.new failed: " .. err)
end
local prefix = "edo-auth.asession:"
local db, err = session_db.new_redis(client, prefix)
if err then
   return test.response_error("new failed: " .. err)
end


local sess = session.new("MLWlc1ICtzbpvKS6ML7EHPYrP2QWM4", "eyJhbGciOiJSUzI1NiIsImtpZCI6IjFlOWdkazcifQ.ewogImlzcyI6ICJodHRwOi8vc2VydmVyLmV4YW1wbGUuY29tIiwKICJzdWIiOiAiMjQ4Mjg5NzYxMDAxIiwKICJhdWQiOiAiczZCaGRSa3F0MyIsCiAibm9uY2UiOiAibi0wUzZfV3pBMk1qIiwKICJleHAiOiAxMzExMjgxOTcwLAogImlhdCI6IDEzMTEyODA5NzAKfQ.ggW8hZ1EuVLuxNuuIJKX_V8a_OMXzR0EHR9R6jgdqrOOF4daGU96Sr_P6qJp6IcmD3HP99Obi1PRs-cwh3LO-p146waJ8IhehcwL7F09JdijmBqkvPeB2T9CJNqeGpe-gccMg4vfKjkM8FcGvnzZUN4_KSP0aAp1tOJ1zZwgjxqGByKHiOtX7TpdQyHE5lcMiKPXfEIQILVq0pc_E2DzL7emopWoaoZTF_m0_N0YzFC6g6EJbOEoRoSK5hoDalrcvRYLSrQAZZKflyuVCyixEoV9GfNQC3_osjzw2PAithfubEEBLuVVk4XUVrWOLrLl0nx7RkKU8NXNHq-rvKMzqg")

local err = db:save(sess, 10)
if err then
   return test.response_error("save failed: " .. err)
end

local sess2, err = db:get(sess:get_id())
if err then
   return test.response_error("get failed: " .. err)
elseif sess2 ~= sess then
   return test.response_error("session is " .. tutil.to_string(sess2) .. " not " .. tutil.to_string(sess))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
