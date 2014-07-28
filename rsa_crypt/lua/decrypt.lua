package.path = package.path..";"..ngx.var.lua_scripts_dir.."/?.lua"

local self_uuid = "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"

local request_headers = ngx.req.get_headers()
local private_key_uuid = request_headers["X-EDO-Private-Key-UUID"]
local encoded_token = request_headers["X-EDO-Auth-Encoded-Token"]

ngx.log(ngx.DEBUG, "key: *** "..private_key_uuid.." ***")
ngx.log(ngx.DEBUG, "token: *** "..encoded_token.." ***")

local public_key_manager = require "public_key_manager"
public_key_manager.get(private_key_uuid)

local openssl = require "openssl"

local public_key_path = "/path/to/public_keys/directory"
local decoded_token = openssl.rsa.base64_verify(encoded_token, public_key_path.."/"..private_key_uuid)

ngx.log(ngx.DEBUG, "decoded: *** "..decoded_token.." ***")

local _, _, sender_uuid, receiver_uuid, timestamp = decoded_token:find("(.*),(.*),(.*)")

ngx.log(ngx.DEBUG, "decoded: *** "..sender_uuid.." ***")
ngx.log(ngx.DEBUG, "decoded: *** "..receiver_uuid.." ***")
ngx.log(ngx.DEBUG, "decoded: *** "..timestamp.." ***")

local local_unix_timestamp = os.date("%s")

if sender_uuid == nil then
   ngx.log(ngx.DEBUG, "sender_uuid is null: invalid format")
   ngx.header["X-EDO-Error-Message"] = "invalid format X-EDO-Auth-Encoded-Token"
   ngx.exit(401)
elseif receiver_uuid ~= self_uuid then
   ngx.log(ngx.DEBUG, "receiver_uuid is invalid")
   ngx.header["X-EDO-Error-Message"] = "invalid receiver"
   ngx.exit(401)
elseif math.abs(local_unix_timestamp - timestamp) > 300 then
   ngx.log(ngx.DEBUG, "timestamp")
   ngx.header["X-EDO-Error-Message"] = "timestamp"
   ngx.exit(401)
else
   ngx.req.set_header("X-EDO-Sender-UUID", sender_uuid)
   ngx.req.set_header("X-EDO-Receiver-UUID", receiver_uuid)
   ngx.req.set_header("X-EDO-Timestamp", timestamp)
end
