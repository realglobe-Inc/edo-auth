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

local redis = require("resty.redis")


-- redis ドライバのラッパー。

-- address   -- redis サーバーのアドレス。
-- keepalive -- 接続をプールしておく期間。ミリ秒。
-- pool_size -- 接続をプールする数。
-- base      -- 接続。

-- メソッド定義。
local redis_wrapper = {

   -- 接続またはプールから取り出す。
   -- retuen: ok, err
   connect = function(self)
      return self.base:connect(self.address)
   end,

   -- 接続をプールする。
   -- retuen: ok, err
   close = function(self)
      return self.base:set_keepalive(self.keepalive, self.pool_size)
   end

}


-- redis ドライバのラッパーを作成する。
-- timeout: 応答待ちの制限時間。ミリ秒。
local new = function(address, timeout, keepalive, pool_size)
   local base, err = redis:new()
   if err then
      return nil, err
   end
   base:set_timeout(timeout)

   local obj = {
      address = address,
      keepalive = keepalive,
      pool_size = pool_size,
      base = base,
   }
   setmetatable(obj, {__index = redis_wrapper})
   return obj
end

return {
   new = new,
}
