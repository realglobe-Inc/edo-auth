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


-- get と save だけの DB。

-- メソッド定義。
local db_base = {

   -- 取得。
   get = function(self, key)
      local _, err = self.client:connect()
      if err then
         return nil, err
      end

      local buff, err = self.client.redis:get(self.prefix .. key)
      self.client:close()
      if err then
         return nil, err
      elseif buff == ngx.null then
         -- 無かった。
         return nil
      end

      -- あった。

      local val, err = cjson.decode(buff)
      if err then
         return nil, err
      end
      return val
   end,

   -- 保存。
   save = function(self, key, val, exp)
      if exp <= 0 then
         return
      end

      local buff, err = cjson.encode(val)
      if err then
         return err
      end

      local _, err = self.client:connect()
      if err then
         return err
      end

      local _, err = self.client.redis:set(self.prefix .. key, buff, "EX", exp)
      self.client:close()
      if err then
         return err
      end
   end,

}


new_base = function(client, prefix)
   local obj = {
      client = client,
      prefix = prefix,
   }
   setmetatable(obj, {__index = db_base})
   return obj
end


return {
   new_base = new_base,
}
