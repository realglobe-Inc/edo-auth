local function log(level, ...)
   local params = {...}
   local str = "*************************"
   for _, val in pairs(params) do
      str = str.." "..tostring(val)
   end
   str = str.." *************************"
   ngx.log(level, str)
end

return {
   err = function(...)
      log(ngx.ERR, ...)
   end,
   warn = function(...)
      log(ngx.WARN, ...)
   end,
   notice = function(...)
      log(ngx.NOTICE, ...)
   end,
   info = function(...)
      log(ngx.INFO, ...)
   end,
   debug = function(...)
      log(ngx.DEBUG, ...)
   end
}
