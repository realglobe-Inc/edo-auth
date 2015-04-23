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

local redis = require("lib.redis")
local test = require("test.test")
local tutil = require("lib.table")
local session = require("lib.ta_session")
local session_db = require("lib.ta_session_db")


-- redis_host: redis のホスト名。
-- redis_port: redis のポート番号。
-- 成功したら 200 OK を返す。

local client, err = redis.new(ngx.var.redis_host, ngx.var.redis_port, 1000, 10 * 1000, 16)
if err then
   return test.response_error("redis.new failed: " .. err)
end
local prefix = "edo-auth.ta-session:"
local db, err = session_db.new_redis(client, prefix)
if err then
   return test.response_error("new failed: " .. err)
end


local sess = session.new("MLWlc1ICtzbpvKS6ML7EHPYrP2QWM4", ngx.time() + 24 * 3600, "OPf7AqiCV6",
                         session.new_access_token("UhotYNPpXV2gw_4T4aTPRdEeZ1M7C3", "jWBVWDzXJU", ngx.time() + 3600),
                         {iss = "http://idp.example.org", sub = "KR1QiTN7swN17ga4"},
                         "https://ta.example.org")

local err = db:save(sess)
if err then
   return test.response_error("save failed: " .. err)
end

local sess2, err = db:get(sess:get_id())
if err then
   return test.response_error("get failed: " .. err)
elseif sess2 ~= sess then
   return test.response_error("session is " .. tutil.to_string(sess2:to_table()) .. " not " .. tutil.to_string(sess:to_table()))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
