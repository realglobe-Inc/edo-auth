math.randomseed(os.time())

return {
   generate = function(length)
      local numbers = "0123456789"
      local upcase_alphabets = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
      local downcase_alphabets = "abcdefghijklmnopqrstuvwxyz"
      local alphanum = numbers..upcase_alphabets..downcase_alphabets
      local return_string = ""
      for i = 1, length, 1 do
         local rand_num = math.random(string.len(alphanum))
         local byte_char = string.byte(alphanum, rand_num)
         return_string = return_string..string.char(byte_char)
      end
      return return_string
   end
}
