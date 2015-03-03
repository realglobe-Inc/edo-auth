-- 秘密鍵の取り扱い。

local logger = require("logger")
local redis = require("redis")

-- どう管理すれば安全なのか知らないので、redis にキャッシュしてしまう。てきとう。
local redis_client = redis:new()

local function get(private_key_uuid, private_key_directory)
   if not private_key_uuid then
      return nil, {message = "no private key uuid"}
   end
   logger.debug("private_key_manager.lua", "get " .. private_key_uuid)

   -- redis を調べる。
   local private_key_string = redis_client:get("private_key:" .. private_key_uuid)
   if private_key_string then
      -- redis にあった。
      logger.debug("private_key_manager.lua", "get " .. private_key_uuid .. " from redis")
      return private_key_string
   end

   if not private_key_directory then
      return nil, {message = "no private key directory"}
   end

   local private_key_file = private_key_directory .. "/" .. private_key_uuid .. ".pem"
   local fd, err = io.open(private_key_file)
   if err then
      return nil, {message = err}
   end
   private_key_string = fd:read("*a")
   fd:close()

   if (not private_key_string) or private_key_string == "" then
      return nil, {message = "no private key in " .. private_key_file}
   end

   -- ファイルであった。
   logger.debug("private_key_manager.lua", "get " .. private_key_uuid .. " from file")

   local cache_expire = 3600 -- TODO てきとう。
   if redis_client:setex("private_key:" .. private_key_uuid, private_key_string, cache_expire) then
      logger.warn("private_key_manager.lua", "caching private key " .. private_key_uuid .. " in redis failed")
   end

   return private_key_string
end

return {
   get = get
}
