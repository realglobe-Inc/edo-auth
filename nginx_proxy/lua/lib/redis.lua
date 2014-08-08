local config = require "config"
local logger = require "logger"
local redis = require "resty.redis"

local client

local function initialize()
   client = redis:new()
   local timeout = tonumber(config.redis.connection_timeout_millisec)
   if timeout then
      logger.debug("redis.lua", "set_timeout:", timeout)
      client:set_timeout(timeout)
   end
   local _, err = client:connect(config.redis.unix_socket_path)
   if err then
      logger.err("redis.lua", "initialize:", err)
      return
   end
end

local function close()
   local _, err = client:close()
   if err then
      logger.err("redis.lua", "close:", err)
   else
      logger.debug("redis.lua", "redis_client closed")
   end
end

return {
   get = function(key)
      logger.debug("redis.lua", "get: key =", key)
      initialize()
      local value, err = client:get(key)
      close()
      if err then
         logger.err("redis.lua", "get:", err)
      elseif value == ngx.null then
         logger.debug("redis.lua", "return ngx.null")
      else
         logger.debug("redis.lua", "value =", value)
         return value
      end
   end,
   setex = function(key, value, expire)
      logger.debug("redis.lua", "setex: key =", key)
      logger.debug("redis.lua", "setex: value =", value)
      logger.debug("redis.lua", "setex: expire =", expire)
      initialize()
      local _, err = client:setex(key, expire, value)
      close()
      if err then
         logger.err("redis.lua", "setex:", err)
      else
         return true
      end
   end
}
