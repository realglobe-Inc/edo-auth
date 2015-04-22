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
local token = require("lib.access_token")


-- 成功したら 200 OK を返す。

local id = "UhotYNPpXV2gw_4T4aTPRdEeZ1M7C3"
local tag = "jWBVWDzXJU"
local exp = ngx.time() + 3600
local idp = "https://idp.example.org"
local scop = {
   ["openid"] = true,
   ["email"] = true,
}
local tok = token.new(id, tag, exp, idp, scop)
if tok:get_id() ~= id then
   return test.response_error("id is " .. tok:get_id() .. " not " .. id)
elseif tok:get_tag() ~= tag then
   return test.response_error("tag is " .. tok:get_tag() .. " not " .. tag)
elseif tok:get_expires_in() ~= exp then
   return test.response_error("expiration date is " .. tok:get_expires_in() .. " not " .. exp)
elseif tok:get_id_provider() ~= idp then
   return test.response_error("ID provider is " .. tok:get_id_provider() .. " not " .. idp)
elseif not tutil.equal(tok:get_scope(), scop) then
   return test.response_error("scope is " .. tutil.to_string(tok:get_scope()) .. " not " .. tutil.to_string(scop))
end

local tok2 = token.new(id, tag, exp, idp, scop)
if tok2 ~= tok then
   return test.response_error("failed to equal " .. tutil.to_string(tok2:to_table()) .. " not " .. tutil.to_string(tok:to_table()))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
