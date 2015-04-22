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
local session = require("lib.user_session")


-- 成功したら 200 OK を返す。

local id = "MLWlc1ICtzbpvKS6ML7EHPYrP2QWM4"
local exp = ngx.time() + 24 * 3600
local sess = session.new(id, exp)
if sess:get_id() ~= id then
   return test.response_error("id is " .. sess:get_id() .. " not " .. id)
elseif sess:get_expires_in() ~= exp then
   return test.response_error("expiration date is " .. sess:get_expires_in() .. " not " .. exp)
elseif sess:get_access_token() then
   return test.response_error("access token is not nil")
elseif sess:get_account() then
   return test.response_error("account is not nil")
elseif sess:get_request() then
   return test.response_error("request is not nil")
elseif sess:get_state() then
   return test.response_error("state is not nil")
elseif sess:get_nonce() then
   return test.response_error("nonce is not nil")
end

local tok_id = "UhotYNPpXV2gw_4T4aTPRdEeZ1M7C3"
local tok_tag = "jWBVWDzXJU"
local tok_exp = ngx.time() + 3600
local tok = session.new_access_token(tok_id, tok_tag, tok_exp)
if tok:get_id() ~= tok_id then
   return test.response_error("token id is " .. tok:get_id() .. " not " .. tok_id)
elseif tok:get_tag() ~= tok_tag then
   return test.response_error("token tag is " .. tok:get_tag() .. " not " .. tok_tag)
elseif tok:get_expires_in() ~= tok_exp then
   return test.response_error("token expiration date is " .. tok:get_expires_in() .. " not " .. tok_exp)
end

local acnt = {
   iss = "http://idp.example.org",
   sub = "KR1QiTN7swN17ga4",
}
local req = "/ui/index.html"
local stat = "ZhcgfcW7VB"
local nonc = "iqus5DRs3H"
sess:set_access_token(tok)
sess:set_account(acnt)
sess:set_request(req)
sess:set_state(stat)
sess:set_nonce(nonc)

if sess:get_access_token() ~= tok then
   return test.response_error("access token is " .. tutil.to_string(sess:get_access_token():to_table()) .. " not " .. tutil.to_string(tok:to_table()))
elseif not tutil.equal(sess:get_account(), acnt) then
   return test.response_error("account is " .. tutil.to_string(sess:get_account()) .. " not " .. tutil.to_string(acnt))
elseif sess:get_request() ~= req then
   return test.response_error("request is " .. sess:get_request() .. " not " .. req)
elseif sess:get_state() ~= stat then
   return test.response_error("state is " .. sess:get_state() .. " not " .. stat)
elseif sess:get_nonce() ~= nonc then
   return test.response_error("nonce is " .. sess:get_nonce() .. " not " .. nonc)
end

local sess2 = session.new(id, exp)
sess2:set_access_token(tok)
sess2:set_account(acnt)
sess2:set_request(req)
sess2:set_state(stat)
sess2:set_nonce(nonc)

if sess2 ~= sess then
   return test.response_error("failed to equal " .. tutil.to_string(sess2:to_table()) .. " not " .. tutil.to_string(sess:to_table()))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
