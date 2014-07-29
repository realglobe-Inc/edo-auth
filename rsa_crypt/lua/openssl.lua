executer = require "executer"

return {
   hmac = function(secret, data)
      return executer.execute("printf "..data.." | openssl dgst -hmac "..secret.." -sha256")
   end,
   bkf = {
      private_keygen = function(private_key_path)
         executer.execute("openssl genrsa 2048 > "..private_key_path)
      end,
      public_keygen = function(private_key_path, public_key_path)
         executer.execute("openssl rsa -pubout < "..private_key_path.." > "..public_key_path)
      end
   },
   rsa = {
      sign = function(data, private_key_path)
         return executer.execute("printf "..data.." | openssl rsautl -sign -inkey "..private_key_path)
      end,
      base64_sign = function(data, private_key_path)
         return executer.execute("printf "..data.." | openssl rsautl -sign -inkey "..private_key_path.." | openssl base64 | tr -d '\\n'")
      end,
      verify = function(data, public_key_path)
         return executer.execute("printf "..data.." | openssl rsautl -verify -pubin -inkey "..public_key_path)
      end,
      base64_verify = function(data, public_key_path)
         return executer.execute("printf "..data.." | openssl base64 -d -A | openssl rsautl -verify -pubin -inkey "..public_key_path)
      end,
      encrypt = function(data, public_key_path)
         return executer.execute("printf "..data.." | openssl rsautl -encrypt -pubin -inkey "..public_key_path)
      end,
      decrypt = function(data, private_key_path)
         return executer.execute("printf "..data.." | openssl rsautl -decrypt -inkey "..private_key_path)
      end,
      sign_hash = function(data, private_key_path, hash_function)
         -- hash_function = md5, sha1, sha256, ..
         return executer.execute("printf "..data.." | openssl dgst -sign "..private_key_path.." -"..hash_function)
      end,
      verify_hash = function(raw_data, encrypted_data, public_key_path, hash_function)
         local tmp_file_path = "/path/to/tmp/file"
         executer.execute("printf "..encrypted_data.." | openssl base64 -d -A > "..tmp_file_path)
         return os.execute("printf "..raw_data.." | openssl dgst -"..hash_function.." -verify "..public_key_path.." -signature "..tmp_file_path)
      end
   }
}
