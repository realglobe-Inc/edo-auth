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
local id_provider = require("lib.id_provider")
local id_provider_db = require("lib.id_provider_db")


-- idp_dir: ID プロバイダ情報が JSON で置いてあるディレクトリ。
-- redis_host: redis のホスト名。
-- redis_port: redis のポート番号。
-- 成功したら 200 OK を返す。
-- 事前に idp に該当する ID プロバイダを登録しておく必要あり。

local client, err = redis.new(ngx.var.redis_host, ngx.var.redis_port, 1000, 10 * 1000, 16)
if err then
   return test.response_error("redis.new failed: " .. err)
end
local prefix = "edo-auth.id_provider:"
local db, err = id_provider_db.new_file(ngx.var.idp_dir, client, prefix, 1000)
if err then
   return test.response_error("new failed: " .. err)
end

local idp = id_provider.new("https://idp.example.org",
                            {
                               [""] = {
                                  ["kty"] = "EC-pem-pub",
                                  ["b"] = [[-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE3tfF/QYgrjnyDzRPycEyx0yZUvX2
xZS8JFQb74c91Oi5OtThEZDqiyltctMoRBmc1JBq9Doh5ZybUQio1aV46A==
-----END PUBLIC KEY-----]],
                               },
                            },
                            "https://idp.example.org/token",
                            "https://idp.example.org/userinfo")

local idp2, err = db:get(idp:get_id())
if err then
   return test.response_error("get failed: " .. err)
elseif idp2 ~= idp then
   return test.response_error("id provider is " .. tutil.to_string(idp2:to_table()) .. " not " .. tutil.to_string(idp:to_table()))
end

local idp3, err = db:get(idp:get_id())
if err then
   return test.response_error("get failed: " .. err)
elseif idp3 ~= idp then
   return test.response_error("cached id provider is " .. tutil.to_string(idp3:to_table()) .. " not " .. tutil.to_string(idp:to_table()))
end

ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
