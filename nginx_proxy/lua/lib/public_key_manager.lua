require "cjson"
local curl_wrapper = require "curl_wrapper"
local logger = require "logger"

-- TODO
local public_key_endpoint = "https://dyqjbmx-ajj-app001.c4sa.net/api/public_keys/get"
local public_key_dir = "/home/c4sa_www/edo-auth/nginx_proxy/tmp/public_keys/"

local function public_key_file_path(key_uuid)
   return public_key_dir..key_uuid
end

local function download(key_uuid)
   local params = {
      auth_key = "uGtY9kA5DLqdA8IpppzeSKKEKKbt0yfr",
      public_key_uuid = key_uuid
   }
   local headers = {
      "Content-Type: application/json"
   }
   curl_wrapper.set_headers(headers)
   local response = curl_wrapper.post(public_key_endpoint, cjson.encode(params))
   logger.debug("public_key_manager.lua", "response.body:", response.body)
   if response.body then
      local response_object = cjson.decode(response.body)
      if not response_object["data"] then
         logger.debug("public_key_manager.lua", "response_object.data not found")
         return
      end
      public_key_string = response_object["data"]["public_key"]
      local fd, err = io.open(public_key_file_path(key_uuid), "w")
      if fd then
         fd:write(public_key_string)
         fd:close()
      else
         logger.debug("public_key_manager.lua", "download:", err)
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
