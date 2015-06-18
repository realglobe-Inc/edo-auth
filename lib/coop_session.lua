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


-- TA セッション。

-- メソッド定義。
local session = {

   -- ID を返す。
   get_id = function(self)
      return self.id
   end,

   -- アカウント情報を返す。
   get_account = function(self)
      return self.acnt
   end,

   -- アカウントタグを返す。
   get_account_tag = function(self)
      return self.acnt_tag
   end,

   -- 連携元 TA を返す。
   get_from_ta = function(self)
      return self.from_ta
   end,
}

local equal = function(o1, o2)
   return o1.id == o2.id
      and o1.acnt == o2.acnt
      and o1.acnt_tag == o2.acnt_tag
      and o1.from_ta == o2.from_ta
end

-- 指定した ID とアカウント情報でユーザーセッションを作成する。
local new = function(id, acnt, acnt_tag, from_ta)
   local obj = {
      id = id,
      acnt = acnt,
      acnt_tag = acnt_tag,
      from_ta = from_ta,
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
