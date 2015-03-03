ngx = {
   status = 0,
   header = {
   },
   body = "",
   var = {
      lua_lib_dir = "./lib"
   },
   req = {
      get_headers = function()
         return ngx.header
      end,
      set_header = function(key, value)
         ngx.header[key] = value
      end,
      get_uri_args = function()
         return {}
      end
   },
   time = function()
      return os.time()
   end,
   cookie_time = function(time)
      return os.date("%c", time)
   end,
   encode_args = function(args)
      return "encoded_args"
   end,
   decode_args = function(encoded_args)
      return {}
   end,
   decode_base64 = function(base64_encrypted_data)
      local mime_base64 = require "mime_base64"
      return mime_base64.base64_decode(base64_encrypted_data)
   end,
   redirect = function(url)
      ngx.status = 302
      ngx.header["Location"] = url
   end,
   say = function(message)
      ngx.body = ngx.body..message.."\n"
   end,
   exit = function(status_code)
      ngx.status = status_code
      error()
   end
}
