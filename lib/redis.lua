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
--
-- {
--     host,         -- redis サーバーのホスト名。
--     port,         -- redis サーバーのポート番号。
--     idle_timeout, -- 接続をプールしておく期間。ミリ秒。
--     pool_size,    -- 接続をプールする数。
--     redis,        -- 接続。
-- }

-- メソッド定義。
local redis_client = {

   -- 接続またはプールから取り出す。
   -- retuen: ok, err
   connect = function(self)
      return self.redis:connect(self.host, self.port)
   end,

   -- 接続をプールする。
   -- retuen: ok, err
   close = function(self)
      return self.redis:set_keepalive(self.idle_timeout, self.pool_size)
   end

}


-- redis ドライバのラッパーを作成する。
-- timeout: 応答待ちの制限時間。ミリ秒。
local new = function(host, port, timeout, idle_timeout, pool_size)
   local red, err = redis:new()
   if err then
      return nil, err
   end
   red:set_timeout(timeout)

   local obj = {
      host = host,
      port = port,
      idle_timeout = idle_timeout,
      pool_size = pool_size,
      redis = red,
   }
   setmetatable(obj, {__index = redis_client})
   return obj
end

return {
   new = new,
}
