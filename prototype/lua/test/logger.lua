local function log(...)
   local args = {...}
   local str = "*************************"
   for _, val in pairs(args) do
      str = str.." "..tostring(val)
   end
   str = str.." *************************"
   print(str)
end

return {
   err = function(...)
      log(...)
   end,
   debug = function(...)
      log(...)
   end
}
