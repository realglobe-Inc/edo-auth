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
local session = require("lib.user_session")
local redis_db = require("lib.redis_db")


-- セッション DB。
-- バックエンドのデータもこのプログラム専用の前提。

-- メソッド定義。
local db_redis = {

   -- 取得。
   get = function(self, id)
      local val, err = self.db:get(id)
      if err then
         return nil, err
      elseif val == nil then
         return nil
      end
      return session.from_table(val)
   end,

   -- 保存。
   save = function(self, sess)
      return self.db:save(sess:get_id(), sess:to_table(), sess:get_expires_in() - ngx.time())
   end,

}


-- redis ドライバとキーの接頭辞を指定して作成する。
new_redis = function(client, prefix)
   local obj = {
      db = redis_db.new_base(client, prefix),
   }
   setmetatable(obj, {__index = db_redis})
   return obj
end


return {
   new_redis = new_redis,
}
