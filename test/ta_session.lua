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
local tutil = require("lib.table")
local session = require("lib.ta_session")


-- 成功したら 200 OK を返す。

local id = "MLWlc1ICtzbpvKS6ML7EHPYrP2QWM4"
local exp = ngx.time() + 24 * 3600
local tag = "OPf7AqiCV6"
local tok = session.new_access_token("UhotYNPpXV2gw_4T4aTPRdEeZ1M7C3", "jWBVWDzXJU", ngx.time() + 3600)
local acnt = {
   iss = "http://idp.example.org",
   sub = "KR1QiTN7swN17ga4",
}
local ta = "https://ta.example.org"
local sess = session.new(id, exp, tag, tok, acnt, ta)

if sess:get_id() ~= id then
   return test.response_error("id is " .. sess:get_id() .. " not " .. id)
elseif sess:get_expires_in() ~= exp then
   return test.response_error("expiration date is " .. sess:get_expires_in() .. " not " .. exp)
elseif sess:get_tag() ~= tag then
   return test.response_error("tag is " .. sess:get_tag() .. " not " .. tag)
elseif sess:get_access_token() ~= tok then
   return test.response_error("access token is " .. tutil.to_string(sess:get_access_token():to_table()) .. " not " .. tutil.to_string(tok:to_table()))
elseif not tutil.equal(sess:get_account(), acnt) then
   return test.response_error("account is " .. tutil.to_string(sess:get_account()) .. " not " .. tutil.to_string(acnt))
elseif sess:get_ta_from() ~= ta then
   return test.response_error("ta from is " .. sess:get_ta_from() .. " not " .. ta)
end

local sess2 = session.new(id, exp, tag, tok, acnt, ta)

if sess2 ~= sess then
   return test.response_error("failed to equal " .. tutil.to_string(sess2:to_table()) .. " not " .. tutil.to_string(sess:to_table()))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
