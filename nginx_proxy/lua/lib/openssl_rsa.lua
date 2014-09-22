require "openssl"

return {
   random_string = function(length)
      local seed = openssl.random(1):byte() * (2 ^ 24)
         + openssl.random(1):byte() * (2 ^ 16)
         + openssl.random(1):byte() * (2 ^ 8)
         + openssl.random(1):byte()
      math.randomseed(seed)

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
   end,
   base64_verify = function(base64_encrypted_data, public_key_string)
      local encrypted_data = ngx.decode_base64(base64_encrypted_data)
      local public_key = openssl.pkey.read(public_key_string, true)
      return public_key:decrypt(encrypted_data)
   end,
   verify_hash = function(raw_data, base64_encrypted_data, public_key_string, hash_function)
      local encrypted_data = ngx.decode_base64(base64_encrypted_data)
      local public_key = openssl.pkey.read(public_key_string, true)
      return public_key:verify(raw_data, encrypted_data, hash_function)
   end,
   sign_base64 = function(data, private_key_string, hash_type)
      hash_type = hash_type or "sha1"
      local private_key = openssl.pkey.read(private_key_string)
      local sign = private_key:sign(data)
      return ngx.encode_base64(sign)
   end
}
