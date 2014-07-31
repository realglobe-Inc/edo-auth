require "cjson"

return {
   get = function(key_uuid)
      if io.open("/home/c4sa_www/edo-auth/nginx_proxy/"..key_uuid) == nil then
         local params = {
            auth_key = "uGtY9kA5DLqdA8IpppzeSKKEKKbt0yfr",
            public_key_uuid = key_uuid
         }
         local options = {
            method = ngx.HTTP_POST,
            body = cjson.encode(params)
         }
         local res = ngx.location.capture("/location/capture/public_keys/get", options)

         ngx.log(ngx.ERR, "*** "..res.body.." ***")

         local response_object = cjson.decode(res.body)
         local public_key_string = response_object["data"]["public_key"]
         local fd = io.open("/home/c4sa_www/edo-auth/nginx_proxy/"..key_uuid, "w")
         fd:write(public_key_string)
         fd:close()
      end
   end
}
