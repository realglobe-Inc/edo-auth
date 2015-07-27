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
local session = require("lib.coop_session")
local session_db = require("lib.coop_session_db")


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


local sess = session.new(
   "MLWlc1ICtzbpvKS6ML7EHPYrP2QWM4",
   "eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MzQ5ODcxMjEsImF0X3RhZyI6Ik5ZbVZnWXNkclUiLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjE2MDQiLCJzdWIiOiJ2Q0FSNkRXbkNxd3VRTHFEU2Z6UzduY05WTDdram5mc1I2OWFEX2pPUU80In0.",
   "reader",
   "https://from.example.org",
   "eyJhbGciOiJub25lIn0.eyJmZ2hpaiI6eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjExNjA0Iiwic3ViIjoiNnR0QmVEaDlNVk56UWs0OF9FTVVGc3BsNEZ2RFFHUU1QQ3BkX1luMllJZyJ9fQ."
)

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
