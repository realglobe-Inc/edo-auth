local redis = require "redis"
local logger = require "logger"

local redis_client = redis:new()

local function get_redis(prefix, key)
   logger.debug("token_manager.lua", "get", "prefix:", prefix, "key:", key)
   if (not prefix) or (not key) then
      return nil
   else
      return redis_client:get(prefix..":"..key)
   end
end

local function set_redis(prefix, key, value, expire)
   logger.debug("token_manager.lua", "set", "prefix:", prefix, "key:", key, "value_length:", value:len(), "expire:", expire)
   if (not prefix) or (not key) or (not value) or (not expire) then
      return nil
   else
      return redis_client:setex(prefix..":"..key, value, expire)
   end
end

local function del_redis(prefix, key)
   logger.debug("token_manager.lua", "del", "prefix:", prefix, "key:", key)
   if (not prefix) or (not key) then
      return nil
   else
      return redis_client:del(prefix..":"..key)
   end
end

return {
   access_token = {
      get = function(key)
         return get_redis("access_token", key)
      end,
      set = function(key, value, expire)
         return set_redis("access_token", key, value, expire)
      end,
      del = function(key)
         return del_redis("access_token", key)
      end
   },
   id_token = {
      get = function(key)
         return get_redis("id_token", key)
      end,
      set = function(key, value, expire)
         return set_redis("id_token", key, value, expire)
      end,
      del = function(key)
         return del_redis("id_token", key)
      end
   }
}
