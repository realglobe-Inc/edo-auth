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
local session = require("lib.coop_session")


-- セッション DB。
-- バックエンドのデータもこのプログラム専用の前提。
-- {
--    "user": <アカウント情報>,
--    "user_tag": <アカウントタグ>,
--    "from_ta": <連携元 TA>
-- }

-- メソッド定義。
local db_redis = {

   -- 取得。
   get = function(self, id)
      local _, err = self.redis:connect()
      if err then
         return nil, err
      end

      local buff, err = self.redis.base:get(self.prefix .. id)
      self.redis:close()
      if err then
         return nil, err
      elseif buff == ngx.null then
         -- 無かった。
         return nil
      end

      local obj, err = cjson.decode(buff)
      if err then
         return nil, err
      end
      return session.new(id, obj.user, obj.user_tag, obj.from_ta, obj.users)
   end,

   -- 保存。
   save = function(self, sess, exp_in)
      if exp_in <= 0 then
         return
      end

      local raw = {
            ["id"] = sess:get_id(),
            ["user"] = sess:get_account(),
            ["user_tag"] = sess:get_account_tag(),
            ["from_ta"] = sess:get_from_ta(),
      }
      local acnts = sess:get_accounts()
      if acnts then
         raw["users"] = acnts
      end
      local buff, err = cjson.encode(raw)
      if err then
         return err
      end

      local _, err = self.redis:connect()
      if err then
         return err
      end

      local _, err = self.redis.base:set(self.prefix .. sess:get_id(), buff, "EX", exp_in)
      self.redis:close()
      if err then
         return err
      end
   end,
}


-- redis ドライバとキーの接頭辞を指定して作成する。
new_redis = function(redis, prefix)
   local obj = {
      redis = redis,
      prefix = prefix,
   }
   setmetatable(obj, {__index = db_redis})
   return obj
end


return {
   new_redis = new_redis,
}
