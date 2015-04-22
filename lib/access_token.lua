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

local tutil = require("lib.table")


-- アクセストークン情報。
--
-- {
--     id,   -- ID。
--     tag,  -- タグ。
--     exp,  -- 有効期限。UNIX 時間。
--     idp,  -- 発行した IdP の ID。
--     scop, -- 許可されているスコープ。スコープ名の集合。
-- }

-- メソッド定義。
local access_token = {

   -- ID を返す。
   get_id = function(self)
      return self.id
   end,

   -- タグ を返す。
   get_tag = function(self)
      return self.tag
   end,

   -- 有効期限を返す。
   get_expires_in = function(self)
      return self.exp
   end,

   -- IdP の ID を返す。
   get_id_provider = function(self)
      return self.idp
   end,

   -- 許可スコープを返す。
   get_scope = function(self)
      return self.scop
   end,

   -- cjson に渡すテーブルにする。
   -- {
   --     "id": <id:文字列>,
   --     "tag": <tag:文字列>,
   --     "expires_in": <exp:数値>,
   --     "issuer": <idp:文字列>,
   --     "scope": <scop:文字列の配列>
   -- }
   to_table = function(self)
      return {
         id = self.id,
         tag = self.tag,
         expires_in = self.exp,
         issuer = self.idp,
         scope = tutil.keys(self.scop),
      }
   end

}

local equal = function(o1, o2)
   return o1.id == o2.id
      and o1.tag == o2.tag
      and o1.exp == o2.exp
      and tutil.equal(o1.scop, o2.scop)
end

-- アクセストークン情報を作成する。
local new = function(id, tag, exp, idp, scop)
   local obj = {
      id = id,
      tag = tag,
      exp = exp,
      idp = idp,
      scop = scop,
   }
   setmetatable(obj, {
                   __index = access_token,
                   __eq = equal,
   })
   return obj
end


-- テーブルからつくる。
local from_table = function(t)
   return new(t.id, t.tag, t.expires_in, t.issuer, tutil.array_to_set(t.scope))
end


return {
   new = new,
   from_table = from_table,
}
