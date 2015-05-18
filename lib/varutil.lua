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


local function get_level(level)
   if (not level) or level == "debug" then
      return ngx.DEBUG
   elseif level == "info" then
      return ngx.INFO
   elseif level == "notice" then
      return ngx.NOTICE
   elseif level == "warn" then
      return ngx.WARN
   elseif level == "error" then
      return ngx.ERR
   elseif level == "crit" then
      return ngx.CRIT
   elseif level == "alert" then
      return ngx.ALERT
   elseif level == "emerg" then
      return ngx.EMERG
   else
      return ngx.DEBUG
   end
end

local function get_boolean(bool, default)
   if not bool then
      return default
   elseif bool == "true" then
      return true
   elseif bool == "false" then
      return false
   else
      return default
   end
end


return {
   get_level = get_level,
   get_boolean = get_boolean,
}
