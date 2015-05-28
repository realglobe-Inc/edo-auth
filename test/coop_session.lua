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
local session = require("lib.coop_session")


-- 成功したら 200 OK を返す。

local id = "MLWlc1ICtzbpvKS6ML7EHPYrP2QWM4"
local acnt = "eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MjY1NjEyNjIsImF0X3RhZyI6InVudG5GZHhOMDMiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsInN1YiI6Ijc1NUI2MjkyMDhFREZEQzIifQ."
local acnt_tag = "reader"
local from_ta = "https://from.example.org"
local sess = session.new(id, acnt, acnt_tag, from_ta)
if sess:get_id() ~= id then
   return test.response_error("id is " .. sess:get_id() .. " not " .. id)
elseif sess:get_account() ~= acnt then
   return test.response_error("account info is " .. sess:get_account() .. " not " .. acnt)
elseif sess:get_account_tag() ~= acnt_tag then
   return test.response_error("account tag is " .. sess:get_account_tag() .. " not " .. acnt_tag)
elseif sess:get_from_ta() ~= from_ta then
   return test.response_error("from-TA is " .. sess:get_from_ta() .. " not " .. from_ta)
end

local sess2 = session.new(id, acnt, acnt_tag, from_ta)
if sess2 ~= sess then
   return test.response_error("failed to equal " .. tutil.to_string(sess2) .. " not " .. tutil.to_string(sess))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
