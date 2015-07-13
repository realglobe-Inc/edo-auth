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


-- 未確立セッション。

-- メソッド定義。
local session = {

   -- ID を返す。
   get_id = function(self)
      return self.id
   end,

   -- 署名用トークンを返す。
   get_token = function(self)
      return self.token
   end,

   -- 通信元アドレスを返す。
   get_address = function(self)
      return self.address
   end,
}

local equal = function(o1, o2)
   return o1.id == o2.id
      and o1.token == o2.token
      and o1.address == o2.address
end

-- 指定した ID とアカウント情報でユーザーセッションを作成する。
local new = function(id, token, address)
   local obj = {
      id = id,
      token = token,
      address = address,
   }
   setmetatable(obj, {
                   __index = session,
                   __eq = equal,
   })
   return obj
end

return {
   new = new,
}
