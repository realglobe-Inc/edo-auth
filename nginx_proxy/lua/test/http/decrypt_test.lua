package.path = package.path..";../?.lua;../lib/?.lua"

local config = require "config"
local rsa_key_uuid = "test_rsa_key_uuid"
config.public_key.public_key_path = config.public_key.local_storage_directory..rsa_key_uuid
config.public_key.private_key_path = config.public_key.public_key_path..".private"
os.execute("openssl genrsa 2048 > "..config.public_key.private_key_path)
os.execute("openssl rsa -pubout < "..config.public_key.private_key_path.." > "..config.public_key.public_key_path)

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         require "decrypt"
      end
   )
)

assert(ngx.status == 401, "X-EDO-Private-Key-UUID ヘッダが無いと 401")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Encoded-Token", "hogehoge")
         require "decrypt"
      end
   )
)

assert(ngx.status == 401, "X-EDO-Auth-Encoded-Token をデコードできないと 401")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}

local timestamp = os.date("%s")
local auth_token = rsa_key_uuid..","..config.service_uuid..","..timestamp
local invalid_uuid_auth_token = rsa_key_uuid..",invalid_uuid,"..timestamp
local invalid_timestamp_auth_token = rsa_key_uuid..","..config.service_uuid..",1000000000"

local command = "printf "..invalid_uuid_auth_token.." | openssl rsautl -sign -inkey "..config.public_key.private_key_path.."| openssl base64 | tr -d '\n'"
local file = io.popen(command, "r")
local encoded_invalid_uuid_auth_token = file:read("*a")
file:close()

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Encoded-Token", encoded_invalid_uuid_auth_token)
         require "decrypt"
      end
   )
)

assert(ngx.status == 401, "トークン内の UUID と自分の UUID が一致しないと 401")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}

local command = "printf "..invalid_timestamp_auth_token.." | openssl rsautl -sign -inkey "..config.public_key.private_key_path.."| openssl base64 | tr -d '\n'"
local file = io.popen(command, "r")
local encoded_invalid_timestamp_auth_token = file:read("*a")
file:close()

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Encoded-Token", encoded_invalid_timestamp_auth_token)
         require "decrypt"
      end
   )
)

assert(ngx.status == 401, "トークンの有効期限が切れていると 401")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}

local command = "printf "..auth_token.." | openssl rsautl -sign -inkey "..config.public_key.private_key_path.."| openssl base64 | tr -d '\n'"
local file = io.popen(command, "r")
local encoded_auth_token = file:read("*a")
file:close()

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Encoded-Token", encoded_auth_token)
         require "decrypt"
      end
   )
)

assert(ngx.header["X-EDO-Sender-UUID"] == rsa_key_uuid, "X-EDO-Sender-UUID がセットされる")
assert(ngx.header["X-EDO-Receiver-UUID"] == config.service_uuid, "X-EDO-Receiver-UUID がセットされる")
assert(ngx.header["X-EDO-Timestamp"] == timestamp, "X-EDO-Timestamp がセットされる")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}



coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Token", auth_token)
         ngx.req.set_header("X-EDO-Auth-Signed-Token", "hogehoge")
         ngx.req.set_header("X-EDO-Hash-Function", "SHA256")
         require "decrypt"
      end
   )
)

assert(ngx.status == 401, "ハッシュ化したトークンを検証できないと 401")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}

local command = "printf "..invalid_uuid_auth_token.." | openssl dgst -sha256 -sign "..config.public_key.private_key_path.."| openssl base64 | tr -d '\n'"
local file = io.popen(command, "r")
local signed_invalid_uuid_auth_token = file:read("*a")
file:close()

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Token", auth_token)
         ngx.req.set_header("X-EDO-Auth-Signed-Token", signed_invalid_uuid_auth_token)
         ngx.req.set_header("X-EDO-Hash-Function", "SHA256")
         require "decrypt"
      end
   )
)

assert(ngx.status == 401, "生のトークンとハッシュ化したトークンが一致しないと 401")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}

local command = "printf "..auth_token.." | openssl dgst -sha256 -sign "..config.public_key.private_key_path.."| openssl base64 | tr -d '\n'"
local file = io.popen(command, "r")
local signed_auth_token = file:read("*a")
file:close()

coroutine.resume(
   coroutine.create(
      function()
         require "nginx_mock"
         ngx.req.set_header("X-EDO-Private-Key-UUID", rsa_key_uuid)
         ngx.req.set_header("X-EDO-Auth-Token", auth_token)
         ngx.req.set_header("X-EDO-Auth-Signed-Token", signed_auth_token)
         require "decrypt"
      end
   )
)

assert(ngx.header["X-EDO-Sender-UUID"] == rsa_key_uuid, "X-EDO-Sender-UUID がセットされる")
assert(ngx.header["X-EDO-Receiver-UUID"] == config.service_uuid, "X-EDO-Receiver-UUID がセットされる")
assert(ngx.header["X-EDO-Timestamp"] == timestamp, "X-EDO-Timestamp がセットされる")
package.loaded["nginx_mock"] = false
package.loaded["decrypt"] = false
ngx.header = {}
