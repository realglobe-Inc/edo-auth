package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local config = require "config"
local logger = require "logger"
local exiter = require "exiter"
local openssl_rsa = require "openssl_rsa"
local public_key_manager = require "public_key_manager"

local request_headers = ngx.req.get_headers()
local private_key_uuid = request_headers["X-EDO-Private-Key-UUID"]
-- private_encrypt
local encoded_token = request_headers["X-EDO-Auth-Encoded-Token"]
-- sign
local raw_token = request_headers["X-EDO-Auth-Token"]
local signed_token = request_headers["X-EDO-Auth-Signed-Token"]
local hash_function = request_headers["X-EDO-Hash-Function"]

if not private_key_uuid then
   exiter.exit("private_key_uuid is blank")
end

logger.debug("decrypt.lua", "private_key_uuid:", private_key_uuid)
local public_key_string = public_key_manager.get(private_key_uuid)

if encoded_token then
   logger.debug("decrypt.lua", "encoded_token:", encoded_token)
   raw_token = openssl_rsa.base64_verify(encoded_token, public_key_string)
   logger.debug("decrypt.lua", "decoded_token:", raw_token)
elseif raw_token and signed_token then
   logger.debug("decrypt.lua", "hash_function:", hash_function)
   hash_function = hash_function or "sha256"
   if not openssl_rsa.verify_hash(raw_token, signed_token, public_key_string, hash_function) then
      exiter.exit("verify failed")
   end
else
   exiter.exit("encoded_token is blank")
end

local _, _, sender_uuid, receiver_uuid, timestamp = raw_token:find("(.*),(.*),(.*)")

logger.debug("decrypt.lua", "sender_uuid:", sender_uuid)
logger.debug("decrypt.lua", "receiver_uuid:", receiver_uuid)
logger.debug("decrypt.lua", "timestamp:", timestamp)

if sender_uuid == nil then
   exiter.exit("invalid format X-EDO-Auth-Encoded-Token")
elseif receiver_uuid ~= config.service_uuid then
   exiter.exit("receiver_uuid is invalid")
elseif config.rsa.token_lifetime then
   local local_unix_timestamp = os.date("%s")
   if math.abs(local_unix_timestamp - timestamp) > config.rsa.token_lifetime then
      exiter.exit("token lifetime is "..config.rsa.token_lifetime.." seconds")
   end
else
   ngx.req.set_header("X-EDO-Sender-UUID", sender_uuid)
   ngx.req.set_header("X-EDO-Receiver-UUID", receiver_uuid)
   ngx.req.set_header("X-EDO-Timestamp", timestamp)

   logger.debug("decrypt.lua", "decrypted")
end
