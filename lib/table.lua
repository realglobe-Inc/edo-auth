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


-- テーブルの比較。
local function equal(t1, t2)
   if t1 == t2 then
      return true
   elseif (not t1) or (not t2) then
      return false
   end

   for k, v1 in pairs(t1) do
      local v2 = t2[k]
      if v1 ~= v2 then
         -- 両方テーブルかつ等しい場合のみ許してやる。
         if type(v1) ~= "table" then
            return false
         elseif type(v2) ~= "table" then
            return false
         elseif not equal(v1, v2) then
            return false
         end
      end
   end

   -- t1 は t2 に含まれる。

   for k, _ in pairs(t2) do
      if not t1[k] then
         return false
      end
   end

   -- t2 は t1 より大きくない。

   return true
end

-- テーブルを文字列にする。
local to_string = function(t)
   local buff, _ = cjson.encode(t)
   if not buff then
      return ""
   end
   return buff
end

-- キーの配列をつくる。
local keys = function(t)
   if not t then
      return nil
   end

   local ks = {}
   local i = 1
   for k, _ in pairs(t) do
      ks[i] = k
      i = i + 1
   end
   return ks
end

-- 配列を集合形式にする。
local array_to_set = function(t)
   if not t then
      return nil
   end

   local set = {}
   for _, v in pairs(t) do
      set[v] = true
   end
   return set
end


return {
   equal = equal,
   to_string = to_string,
   keys = keys,
   array_to_set = array_to_set,
}
