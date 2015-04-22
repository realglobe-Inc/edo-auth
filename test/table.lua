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


-- 成功したら 200 OK を返す。

if not tutil.equal({}, {}) then
   return test.response_error("failed to equal {}")
elseif not tutil.equal({a = "b"}, {a = "b"}) then
   return test.response_error('failed to equal {a = "b"}')
elseif tutil.equal({a = "b"}, {a = "c"}) then
   return test.response_error('failed to distinguish between {a = "b"} and {a = "c"}')
elseif tutil.equal({a = "b"}, {c = "b"}) then
   return test.response_error('failed to distinguish between {a = "b"} and {c = "b"}')
end

ngx.status = ngx.HTTP_OK
return ngx.exit(ngx.status)
