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
local id_provider = require("lib.id_provider")


-- アクセストークン情報 DB。
-- バックエンドのデータを他のプログラムと共有する前提。

-- メソッド定義。
local db_file = {

   -- 取得。
   get = function(self, id)
      local _, err = self.client:connect()
      if err then
         return nil, err
      end

      local buff, _ = self.client.redis:get(self.prefix .. id)
      if buff == ngx.null then
         -- キャッシュされて無かった。

         local base = ngx.escape_uri(id)
         local fd, err = io.open(self.dir .. "/" .. base .. ".json")
         if err then
            -- ファイル無し。
            self.client:close()
            return nil, err
         end

         buff = fd:read("*a")
         fd:close()
      end

      local idp, err = cjson.decode(buff)
      if err then
         self.client:close()
         return nil, err
      end

      -- キャッシュする。
      self.client.redis:set(self.prefix .. id, buff, "EX", self.expires)
      self.client:close()

      return id_provider.from_table(idp)
   end,

}


-- ディレクトリを指定して作成する。
-- client: キャッシュを保存する redis ドライバ。
-- prefix: キャッシュキーの接頭辞。
-- expires: キャッシュの生存期間。ミリ秒。
new_file = function(dir, client, prefix, expires)
   if not (string.len(dir) > 0 and string.sub(dir, 1, 1) == "/") then
      -- 相対パス。
      dir = ngx.config.prefix() .. "/" .. dir
   end

   local obj = {
      dir = dir,
      client = client,
      prefix = prefix,
      expires = expires,
   }
   setmetatable(obj, {__index = db_file})
   return obj
end


return {
   new_file = new_file,
}
