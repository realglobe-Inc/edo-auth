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
local usession = require("lib.user_session")


-- TA セッション。
--
-- {
--     id,   -- セッション ID。
--     exp,  -- 有効期限。UNIX 時間。
--     tag,  -- 処理の主体のアカウントタグ。
--     tok,  -- 処理の主体のアクセストークン。
--     acnt, -- 処理の主体の情報。
--     ta,   -- アクセス元 TA の ID。
-- }

-- メソッド定義。
local session = {

   -- ID を返す。
   get_id = function(self)
      return self.id
   end,

   -- 有効期限を返す。
   get_expires_in = function(self)
      return self.exp
   end,

   -- 処理の主体のアカウントタグを返す。
   get_tag = function(self)
      return self.tag
   end,

   -- 処理の主体のアクセストークンを返す。
   get_access_token = function(self)
      return self.tok
   end,

   -- 処理の主体の情報を返す。
   get_account = function(self)
      return self.acnt
   end,

   -- アクセス元 TA の ID を返す。
   get_ta_from = function(self)
      return self.ta
   end,

   -- cjson に渡すテーブルにする。
   to_table = function(self)
      local tok
      if self.tok then
         tok = self.tok:to_table()
      end
      return {
         id = self.id,
         expires_in = self.exp,
         tag = self.tag,
         access_token = tok,
         account = self.acnt,
         ta_from = self.ta,
      }
   end
}

local equal = function(o1, o2)
   return o1.id == o2.id
      and o1.exp == o2.exp
      and o1.tag == o2.tag
      and o1.tok == o2.tok
      and tutil.equal(o1.acnt, o2.acnt)
      and o1.ta == o2.ta
end

-- セッションを作成する。
local new = function(id, exp, tag, tok, acnt, ta)
   local obj = {
      id = id,
      exp = exp,
      tag = tag,
      tok = tok,
      acnt = acnt,
      ta = ta,
   }
   setmetatable(obj, {
                   __index = session,
                   __eq = equal,
   })
   return obj
end


-- テーブルからつくる。
local from_table = function(t)
   local v = t.access_token
   return new(t.id, t.expires_in, t.tag, usession.new_access_token(v.id, v.tag, v.expires_in),
              t.account, t.ta_from)
end


return {
   new = new,
   new_access_token = usession.new_access_token,
   from_table = from_table,
}
