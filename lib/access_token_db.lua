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

local cjson = require("cjson.safe")
local token = require("lib.access_token")


-- アクセストークン情報 DB。
-- バックエンドのデータを他のプログラムと共有する前提。

-- メソッド定義。
local db_redis = {

   -- 取得。
   -- テスト用。
   get_by_tag = function(self, tag)
      local _, err = self.client:connect()
      if err then
         return nil, err
      end

      local buff, err = self.client.redis:get(self.prefix .. tag)
      self.client:close()
      if err then
         return nil, err
      elseif buff == ngx.null then
         -- 無かった。
         return nil
      end

      -- あった。

      local tok, err = cjson.decode(buff)
      if err then
         return nil, err
      end
      return token.from_table(tok)
   end,

   -- 保存。
   save = function(self, tok)
      -- 以降の時間の掛かり方によってはちょっと長めに期限を定めていることになる。
      local exp = tok:get_expires_in() - ngx.time()
      if exp <= 0 then
         return
      end

      local buff, err = cjson.encode(tok:to_table())
      if err then
         return err
      end

      local _, err = self.client:connect()
      if err then
         return err
      end

      -- タグで保存。
      local _, err = self.client.redis:set(self.prefix .. tok:get_tag(), buff, "EX", exp)
      self.client:close()
      if err then
         return err
      end
   end,

}


-- redis ドライバとキーの接頭辞を指定して作成する。
new_redis = function(client, prefix)
   local obj = {
      client = client,
      prefix = prefix,
   }
   setmetatable(obj, {__index = db_redis})
   return obj
end


return {
   new_redis = new_redis,
}
