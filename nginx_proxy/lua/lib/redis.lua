local logger = require "logger"

local function get_client()
   local redis = require "resty.redis"
   local client = redis:new()
   client:set_timeout(1000)
   local _, err = client:connect("unix:/home/c4sa_www/redis-2.8.12/redis.sock")
   if err then
      logger.debug("redis.lua", "get_client:", err)
      return
   end
   return client
end

local function close_client(client)
   local _, err = client:close()
   if err then
      logger.debug("redis.lua", "close:", err)
   end
end

return {
   get = function(key)
      local client = get_client()
      if not client then
         return
      end
      local value, err = client:get(key)
      close_client(client)
      logger.debug("redis.lua", "key:", key)
      logger.debug("redis.lua", "value:", value)
      if err then
         logger.debug("redis.lua", "get:", err)
      elseif value == ngx.null then
         logger.debug("redis.lua", "return ngx.null")
      else
         return value
      end
   end,
   setex = function(key, value, expire)
      logger.debug("redis.lua", "key:", key)
      logger.debug("redis.lua", "value:", value)
      logger.debug("redis.lua", "expire:", expire)
      local client = get_client()
      if not client then
         return
      end
      local ok, err = client:setex(key, expire, value)
      if err then
         logger.debug("redis.lua", "setex:", err)
      end
      close_client(client)
      return ok
   end
}
