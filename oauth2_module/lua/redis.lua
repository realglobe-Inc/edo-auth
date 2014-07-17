return {
   get = function(key)
      local redis_get_path = ngx.var.local_proxy_namespace.."/redis/get"
      local response = ngx.location.capture(redis_get_path.."/"..key)
      if response.status == 200 then
         _, _, value = string.find(response.body, [[$%d+%s*(.*)%s*]])
         return value
      end
   end,
   setex = function(key, value, expire)
      local redis_setex_path = ngx.var.local_proxy_namespace.."/redis/setex"
      local response = ngx.location.capture(redis_setex_path.."/"..key.."/"..value.."/"..expire)
      return response.status == 200
   end
}
