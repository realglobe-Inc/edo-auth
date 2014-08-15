local json_safe = require "cjson.safe"
local data_file = "redis_mock_data.json"
local logger = require "logger"

local function initialize(self)
   local f, err = io.open(data_file)
   if f then
      local json_string = f:read("*a")
      f:close()
      self.client = json_safe.decode(json_string) or {}
   else
      logger.debug("test/redis.lua", "initialize", err)
      self.client = {}
   end
end

local function close(self)
   local f, err = io.open(data_file, "w")
   if f then
      local json_string = json_safe.encode(self.client)
      f:write(json_string)
      f:close()
   else
      logger.debug("test/redis.lua", "close", f)
   end
end

local function clear(self)
   logger.debug("test/redis.lua", "clear")
   self.client = {}
   self:close()
end

local function get(self, key)
   logger.debug("test/redis.lua", "get", "key:", key)
   self:initialize()
   local value = self.client[key]
   self:close()
   logger.debug("test/redis.lua", "get", "value:", value)
   return value
end

local function setex(self, key, value, expire)
   logger.debug("test/redis.lua", "setex", key, value)
   self:initialize()
   self.client[key] = value
   self:close()
   return true
end

return {
   new = function(self)
      return {
         client,
         initialize = initialize,
         close = close,
         get = get,
         setex = setex,
         clear = clear
      }
   end
}
