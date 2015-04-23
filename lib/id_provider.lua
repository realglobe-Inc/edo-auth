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


-- ID プロバイダ情報。
--
-- {
--     id,        -- ID。
--     veri_keys, -- 検証鍵。kid から JWK を基にしたテーブルへのマップ。
--     tok_uri,   -- トークンエンドポイント。
--     acnt_uri,  -- アカウント情報エンドポイント。
-- }

-- 検証鍵について。
-- 現状ではまともに JWK を扱えないため、RSA-pem-pub と EC-pem-pub 形式を定義する。
-- どちらも PEM 形式の公開鍵を b 要素に入れたものとする。

-- メソッド定義。
local id_provider = {

   -- ID を返す。
   get_id = function(self)
      return self.id
   end,

   -- 検証鍵を返す。
   get_verify_keys = function(self)
      return self.veri_keys
   end,

   -- トークンエンドポイントを返す。
   get_token_uri = function(self)
      return self.tok_uri
   end,

   -- アカウント情報エンドポイントを返す。
   get_account_uri = function(self)
      return self.acnt_uri
   end,

   to_table = function(self)
      return {
         issuer = self.id,
         verify_keys = tutil.values(self.veri_keys),
         token_endpoint = self.tok_uri,
         userinfo_endpoint = self.acnt_uri,
      }
   end,
}

local equal = function(o1, o2)
   return o1.id == o2.id
      and tutil.equal(o1.veri_keys, o2.veri_keys)
      and o1.tok_uri == o2.tok_uri
      and o1.acnt_uri == o2.acnt_uri
end

-- ID プロバイダ情報を作成する。
local new = function(id, veri_keys, tok_uri, acnt_uri)
   local obj = {
      id = id,
      veri_keys = veri_keys,
      tok_uri = tok_uri,
      acnt_uri = acnt_uri,
   }
   setmetatable(obj, {
                   __index = id_provider,
                   __eq = equal,
   })
   return obj
end


local is_verify_jwk = function(key)
   if key.use == "enc" then
      -- 暗号用だった。
      return false
   elseif key.key_ops then
      for _, v in pairs(key.key_ops) do
         if v == "verify" then
            -- 検証用だった。
            return true
         end
      end
      -- 検証用ではなかった。
      return false
   end
   -- 用途が指定されてなかった。
   return key
end

-- テーブルからつくる。
-- keys から検証用の鍵だけを抜き出して verify_keys にする。
local from_table = function(t)
   local veri_keys
   if t.keys then
      for _, key in pairs(t.keys) do
         if is_verify_jwk(key) then
            if not veri_keys then
               veri_keys = {}
            end
            veri_keys[key.kid or ""] = key
         end
      end
   end
   return new(t.issuer, veri_keys, t.token_endpoint, t.userinfo_endpoint)
end


return {
   new = new,
   from_table = from_table,
}
