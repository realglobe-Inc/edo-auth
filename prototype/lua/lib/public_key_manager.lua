local json_safe = require "cjson.safe"
local config = require "config"
local logger = require "logger"
local curl_wrapper = require "curl_wrapper"

local function public_key_file_path(key_uuid)
   return config.public_key.local_storage_directory.."/"..key_uuid
end

local function download(key_uuid)
   local public_key_string = nil
   local params = {
      auth_key = "uGtY9kA5DLqdA8IpppzeSKKEKKbt0yfr",
      public_key_uuid = key_uuid
   }
   local headers = {
      "Content-Type: application/json"
   }
   local curl_client = curl_wrapper:new()
   curl_client:set_headers(headers)
   local response = curl_client:post(config.public_key.api_endpoint, json_safe.encode(params))
   logger.debug("public_key_manager.lua", "response.body:", response.body)
   if response.body then
      local response_object = json_safe.decode(response.body)
      if response_object["data"] then
         public_key_string = response_object["data"]["public_key"]
      elseif response_object["service"] then
         public_key_string = response_object["service"]["public_key"]
      end

      if not public_key_string then
         logger.err("public_key_manager.lua", "response_object.data not found")
         return
      end

      local fd, err = io.open(public_key_file_path(key_uuid), "w")
      if fd then
         fd:write(public_key_string)
         fd:close()
      else
         logger.err("public_key_manager.lua", "download:", err)
      end
   end
   return public_key_string
end

return {
   get = function(key_uuid)
      local public_key_string
      local fd = io.open(public_key_file_path(key_uuid))
      if fd then
         public_key_string = fd:read("*a")
         fd:close()
      else
         public_key_string = download(key_uuid)
      end
      return public_key_string
   end
}
