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
local id_provider = require("lib.id_provider")


-- 成功したら 200 OK を返す。

local id = "UhotYNPpXV2gw_4T4aTPRdEeZ1M7C3"
local veri_keys = {
   [""] = {
      ["kty"] = "EC-pem-pub",
      ["b"] = [[-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE3tfF/QYgrjnyDzRPycEyx0yZUvX2
xZS8JFQb74c91Oi5OtThEZDqiyltctMoRBmc1JBq9Doh5ZybUQio1aV46A==
-----END PUBLIC KEY-----]],
   },
}
local tok_uri = "https://idp.example.org/token"
local acnt_uri = "https://idp.example.org/userinfo"

local idp = id_provider.new(id, veri_keys, tok_uri, acnt_uri)
if idp:get_id() ~= id then
   return test.response_error("id is " .. idp:get_id() .. " not " .. id)
elseif not tutil.equal(idp:get_verify_keys(), veri_keys) then
   return test.response_error("verify keys are " .. tutil.to_string(idp:get_verify_keys()) .. " not " .. tutil.to_string(veri_keys))
elseif idp:get_token_uri() ~= tok_uri then
   return test.response_error("access token uri is " .. idp:get_token_uri() .. " not " .. tok_uri)
elseif idp:get_account_uri() ~= acnt_uri then
   return test.response_error("account info uri is " .. idp:get_account_uri() .. " not " .. acnt_uri)
end

local idp2 = id_provider.new(id, veri_keys, tok_uri, acnt_uri)
if idp2 ~= idp then
   return test.response_error("failed to equal " .. tutil.to_string(idp2:to_table()) .. " not " .. tutil.to_string(idp:to_table()))
end


ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
