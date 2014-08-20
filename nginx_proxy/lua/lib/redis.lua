local config = require "config"
local logger = require "logger"
local redis = require "resty.redis"

local function initialize(self)
   self.client = redis:new()
   local timeout = tonumber(config.redis.connection_timeout_millisec)
   if timeout then
      logger.debug("redis.lua", "set_timeout:", timeout)
      self.client:set_timeout(timeout)
   end
   local _, err = self.client:connect(config.redis.unix_socket_path)
   if err then
      logger.err("redis.lua", "initialize:", err)
      return
   end
end

local function close(self)
   local _, err = self.client:close()
   if err then
      logger.err("redis.lua", "close:", err)
   else
      logger.debug("redis.lua", "redis_client closed")
   end
end

local function get(self, key)
   logger.debug("redis.lua", "get: key =", key)
   initialize(self)
   local value, err = self.client:get(key)
   close(self)
   if err then
      logger.err("redis.lua", "get:", err)
   elseif value == ngx.null then
      logger.debug("redis.lua", "return ngx.null")
   else
      logger.debug("redis.lua", "value =", value)
      return value
   end
end

local function setex(self, key, value, expire)
   logger.debug("redis.lua", "setex: key =", key)
   logger.debug("redis.lua", "setex: value =", value)
   logger.debug("redis.lua", "setex: expire =", expire)
   initialize(self)
   local _, err = self.client:setex(key, expire, value)
   close(self)
   if err then
      logger.err("redis.lua", "setex:", err)
   else
      return true
   end
end

return {
   new = function(self)
      new_object = {
         get = get,
         setex = setex
      }
      return new_object
   end
}
