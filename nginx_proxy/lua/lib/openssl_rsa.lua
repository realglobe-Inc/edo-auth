require "openssl"

return {
   base64_verify = function(base64_encrypted_data, public_key_string)
      local encrypted_data = ngx.decode_base64(base64_encrypted_data)
      local public_key = openssl.pkey.read(public_key_string, true)
      return public_key:decrypt(encrypted_data)
   end,
   verify_hash = function(raw_data, base64_encrypted_data, public_key_string, hash_function)
      local encrypted_data = ngx.decode_base64(base64_encrypted_data)
      local public_key = openssl.pkey.read(public_key_string, true)
      return public_key:verify(raw_data, encrypted_data, hash_function)
   end
}
