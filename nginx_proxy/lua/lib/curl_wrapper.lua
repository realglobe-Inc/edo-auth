require "cURL"
local logger = require "logger"

local client = cURL.easy_init()
local request_headers = {}
local response_code
local response_headers
local response_body
local response
local callback_options = {
   writefunction = function(str)
      response_body = response_body..str
      logger.debug("curl_wrapper.lua", "data_unit_size:", str:len())
      logger.debug("curl_wrapper.lua", "response_body_size:", response_body:len())
   end,
   headerfunction = function(str)
      if not response_code then
         -- HTTP/1.1 200 OK
         local _, _, code_string = str:find("[^ ]* (%d*) .*")
         response_code = tonumber(code_string) or 0
         logger.debug("curl_wrapper.lua", "response_code:", response_code)
      else
         local _, _, key, value = str:find("(.*):%s*([^%c]*)")
         if key then
            logger.debug("curl_wrapper.lua", "response_headers", key..":", value)
            response_headers[key] = value
         end
      end
   end
}

local function initialize(url)
   response_code = nil
   response_headers = {}
   response_body = ""
   client:setopt_url(url)
   client:setopt_verbose(1)
end

local function perform()
   client:setopt_httpheader(request_headers)
   client:perform(callback_options)
   request_headers = {}
   response = {
      code = response_code,
      headers = response_headers,
      body = response_body
   }
end

return {
   set_headers = function(headers)
      request_headers = headers
   end,
   get = function(url)
      logger.debug("curl_wrapper.lua", "url:", url)
      for key, val in pairs(request_headers) do
         logger.debug("curl_wrapper.lua", "headers", "key:", key, "val:", val)
      end
      initialize(url)
      perform()
      return response
   end,
   post = function(url, request_body)
      initialize(url)
      client:setopt_postfields(request_body)
      perform()
      return response
   end
}
