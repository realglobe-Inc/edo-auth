package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local exiter = require "exiter"

-- TODO
local self_uuid = "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
local public_key_directory = "/home/c4sa_www/edo-auth/nginx_proxy/"

local request_headers = ngx.req.get_headers()
local private_key_uuid = request_headers["X-EDO-Private-Key-UUID"]
-- private_encrypt
local encoded_token = request_headers["X-EDO-Auth-Encoded-Token"]
-- sign
local raw_token = request_headers["X-EDO-Auth-Token"]
local signed_token = request_headers["X-EDO-Auth-Signed-Token"]
local hash_function = request_headers["X-EDO-Hash-Function"]

local openssl_rsa = require "openssl_rsa"

if private_key_uuid == nil then
   exiter.exit("private_key_uuid is blank")
else
   ngx.log(ngx.DEBUG, "private_key_uuid: *** "..private_key_uuid.." ***")

   local public_key_manager = require "public_key_manager"
   public_key_manager.get(private_key_uuid)
end

local public_key_path = public_key_directory..private_key_uuid

if encoded_token ~= nil then
   ngx.log(ngx.DEBUG, "token: *** "..encoded_token.." ***")

   raw_token = openssl_rsa.base64_verify(encoded_token, public_key_path)
   ngx.log(ngx.DEBUG, "decoded: *** "..raw_token.." ***")
elseif (raw_token ~= nil) and (signed_token ~= nil) then
   if hash_function == nil then
      hash_function = "sha256"
   end
   if not openssl_rsa.verify_hash(raw_token, signed_token, public_key_path, hash_function) then
      exiter.exit("verify failed")
   end
else
   exiter.exit("encoded_token is blank")
end

local _, _, sender_uuid, receiver_uuid, timestamp = raw_token:find("(.*),(.*),(.*)")

ngx.log(ngx.DEBUG, "decoded: *** "..sender_uuid.." ***")
ngx.log(ngx.DEBUG, "decoded: *** "..receiver_uuid.." ***")
ngx.log(ngx.DEBUG, "decoded: *** "..timestamp.." ***")

local local_unix_timestamp = os.date("%s")

if sender_uuid == nil then
   exiter.exit("invalid format X-EDO-Auth-Encoded-Token")
elseif receiver_uuid ~= self_uuid then
   exiter.exit("receiver_uuid is invalid")
elseif math.abs(local_unix_timestamp - timestamp) > 300 then
   exiter.exit("timestamp is over 300")
else
   ngx.req.set_header("X-EDO-Sender-UUID", sender_uuid)
   ngx.req.set_header("X-EDO-Receiver-UUID", receiver_uuid)
   ngx.req.set_header("X-EDO-Timestamp", timestamp)

   ngx.log(ngx.DEBUG, "decrypted")
end
