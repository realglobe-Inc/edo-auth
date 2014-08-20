require "cURL"
local logger = require "logger"

local function initialize(self)
   local function writefunction(str)
      self.response_body = self.response_body..str
      logger.debug("curl_wrapper.lua", "data_unit_size:", str:len())
      logger.debug("curl_wrapper.lua", "response_body_size:", self.response_body:len())
   end
   local function headerfunction(str)
      if not self.response_code then
         -- HTTP/1.1 200 OK
         local _, _, code_string = str:find("[^ ]* (%d*) .*")
         self.response_code = tonumber(code_string) or 0
         logger.debug("curl_wrapper.lua", "response_code:", self.response_code)
      else
         local _, _, key, value = str:find("(.*):%s*([^%c]*)")
         if key then
            logger.debug("curl_wrapper.lua", "response_headers", key..":", value)
            self.response_headers[key] = value
         end
      end
   end

   self.client = cURL.easy_init()
   self.request_headers = {}
   self.response_code = nil
   self.response_headers = {}
   self.response_body = ""
   self.client:setopt_verbose(1)
   self.callback_options = {
      writefunction = writefunction,
      headerfunction = headerfunction
   }
end

local function set_url(self, url)
   logger.debug("curl_wrapper.lua", "url:", url)
   for key, val in pairs(self.request_headers) do
      logger.debug("curl_wrapper.lua", "headers", "key:", key, "val:", val)
   end
   self.client:setopt_url(url)
end

local function perform(self)
   self.client:setopt_httpheader(self.request_headers)
   self.client:perform(self.callback_options)
   self.response = {
      code = self.response_code,
      headers = self.response_headers,
      body = self.response_body
   }
   initialize(self)
end

local function set_headers(self, headers)
   self.request_headers = headers
end

local function get(self, url)
   logger.debug("curl_wrapper.lua", "GET")
   set_url(self, url)
   perform(self)
   return self.response
end

local function post(self, url, request_body)
   logger.debug("curl_wrapper.lua", "POST")
   set_url(self, url)
   self.client:setopt_postfields(request_body)
   perform(self)
   return self.response
end

return {
   new = function(self)
      new_object = {
         set_headers = set_headers,
         get = get,
         post = post
      }
      initialize(new_object)
      return new_object
   end
}
