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


-- セッション。
--
-- {
--     id,   -- セッション ID。
--     exp,  -- 有効期限。UNIX 時間。
--     tok,  -- ユーザー認証にて発行されたアクセストークンの情報。
--     acnt, -- ユーザー認証にて得られたアカウント情報。
--     req,  -- ユーザー認証前のリクエスト内容。
--     stat, -- ユーザー認証用 state。
--     nonc, -- ユーザー認証用 nonce。
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

   -- 紐付くアクセストークンを返す。
   get_access_token = function(self)
      return self.tok
   end,

   set_access_token = function(self, tok)
      self.tok = tok
   end,

   -- 紐付くアカウント情報を返す。
   get_account = function(self)
      return self.acnt
   end,

   set_account = function(self, acnt)
      self.acnt = acnt
   end,

   -- 紐付くリクエスト内容を返す。
   get_request = function(self)
      return self.req
   end,

   set_request = function(self, req)
      self.req = req
   end,

   -- 紐付く state 値を返す。
   get_state = function(self)
      return self.stat
   end,

   set_state = function(self, stat)
      self.stat = stat
   end,

   -- 紐付く nonce 値を返す。
   get_nonce = function(self)
      return self.nonc
   end,

   set_nonce = function(self, nonc)
      self.nonc = nonc
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
         access_token = tok,
         account = self.acnt,
         request = self.req,
         state = self.stat,
         nonce = self.nonc,
      }
   end
}

local equal = function(o1, o2)
   return o1.id == o2.id
      and o1.exp == o2.exp
      and o1.tok == o2.tok
      and tutil.equal(o1.acnt, o2.acnt)
      and o1.req == o2.req
      and o1.stat == o2.stat
      and o1.nonc == o2.nonc
end

-- 指定した ID と有効期限でセッションを作成する。
local new = function(id, exp)
   local obj = {
      id = id,
      exp = exp,
   }
   setmetatable(obj, {
                   __index = session,
                   __eq = equal,
   })
   return obj
end


-- セッションに紐付けるアクセストークン情報。
--
-- {
--     id,         -- ID。
--     tag,        -- タグ
--     exp, -- 有効期限。
-- }

-- メソッド定義。
local access_token = {

   -- ID を返す。
   get_id = function(self)
      return self.id
   end,

   -- タグを返す。
   get_tag = function(self)
      return self.tag
   end,

   -- 有効期限を返す。
   get_expires_in = function(self)
      return self.exp
   end,

   -- cjson に渡すテーブルにする。
   to_table = function(self)
      return {
         id = self.id,
         tag = self.tag,
         expires_in = self.exp,
      }
   end,
}


local equal_access_token = function(o1, o2)
   return o1.id == o2.id
      and o1.tag == o2.tag
      and o1.exp == o2.exp
end

-- 指定した ID、タグ、有効期限でアクセストークン情報を作成する。
local new_access_token = function(id, tag, exp)
   local obj = {
      id = id,
      tag = tag,
      exp = exp,
   }
   setmetatable(obj, {
                   __index = access_token,
                   __eq = equal_access_token,
   })
   return obj
end


-- テーブルからつくる。
local from_table = function(t)
   local sess = new(t.id, t.expires_in)
   local v = t.access_token
   if v then
      sess:set_access_token(new_access_token(v.id, v.tag, v.expires_in))
   end
   sess:set_account(t.account)
   sess:set_request(t.request)
   sess:set_state(t.state)
   sess:set_nonce(t.nonce)
   return sess
end


return {
   new = new,
   new_access_token = new_access_token,
   from_table = from_table,
}
