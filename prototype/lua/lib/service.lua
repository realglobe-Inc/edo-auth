-- サービス認証。

local logger = require("logger")
local openssl_rsa = require("openssl_rsa")

-- 返り値は、
-- auth_info
-- {
--   sender_uuid = <送り側サービスの UUID>,
--   receiver_uuid = <受け取り側サービス、つまり自分の UUID>,
--   timestamp = <署名日時>
-- }
local function post_verification(raw_token, my_uuid, token_lifetime)
   local _, _, sender_uuid, receiver_uuid, timestamp = raw_token:find("(.*),(.*),(.*)")
   if not sender_uuid then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "no sender uuid"}
   elseif receiver_uuid ~= my_uuid then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "receiver uuid is wrong"}
   elseif not timestamp then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "no time stamp"}
   end

   local ts = tonumber(timestamp)

   if token_lifetime then
      local local_timestamp = os.date("%s")
      if math.abs(tonumber(local_timestamp) - ts) > token_lifetime then
         return nil, {status = ngx.HTTP_FORBIDDEN, message = "too large gap between encoded token time " .. timestamp .. " and current time " .. local_timestamp}
      end
   end

   logger.debug("decrypt.lua", "service was authorized")

   return {sender_uuid = sender_uuid, receiver_uuid = receiver_uuid, timestamp = ts}
end

local function del_header()
   ngx.req.clear_header("X-EDO-Private-Key-UUID")
end

local function add_header(auth_info)
   ngx.req.set_header("X-EDO-Sender-UUID", auth_info.sender_uuid)
   ngx.req.set_header("X-EDO-Receiver-UUID", auth_info.receiver_uuid)
   ngx.req.set_header("X-EDO-Timestamp", auth_info.timestamp)
end

-- サービス認証する。直接暗号方式。
local function authorize(public_key_string, encoded_token, my_uuid, token_lifetime)
   local raw_token = openssl_rsa.base64_verify(encoded_token, public_key_string)

   if not raw_token then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "cannot decode encoded token"}
   end

   logger.debug("decrypt.lua", "decoded token is " .. raw_token)

   return post_verification(raw_token, my_uuid, token_lifetime)
end

-- サービス認証済みであることをヘッダに示す。直接暗号方式用。
local function set_header(auth_info)
   del_header()
   ngx.req.clear_header("X-EDO-Auth-Encoded-Token")
   add_header(auth_info)
end

-- サービス認証する。ハッシュ値方式。
local function authorize_sign(public_key_string, raw_token, signed_token, hash_function, my_uuid, token_lifetime)
   logger.debug("decrypt.lua", "raw token is " .. raw_token)

   hash_function = hash_function or "sha256"
   if not openssl_rsa.verify_hash(raw_token, signed_token, public_key_string, hash_function) then
      return nil, {status = ngx.HTTP_FORBIDDEN, message = "verifying signed token failed"}
   end

   return post_verification(raw_token, my_uuid, token_lifetime)
end

-- サービス認証済みであることをヘッダに示す。ハッシュ値方式用。
local function set_header_sign(auth_info)
   del_header()
   ngx.req.clear_header("X-EDO-Auth-Token")
   ngx.req.clear_header("X-EDO-Auth-Signed-Token")
   ngx.req.clear_header("X-EDO-Hash-Function")
   add_header(auth_info)
end

return {
   authorize = authorize,
   set_header = set_header,
   authorize_sign = authorize_sign,
   set_header_sign = set_header_sign
}
