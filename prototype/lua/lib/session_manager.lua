-- セッション管理。

local json_safe = require "cjson.safe" -- table に対応してない？みたいだから。
local redis = require("redis")
local logger = require("logger")
local openssl_rsa = require("openssl_rsa")

-- session
-- {
--   key = <セッション ID>,
--   access_token = <アクセストークン>,
--   expires = <有効期限 (期間じゃない)>,
--   id_provider_uuid = <アクセストークンを発行した ID プロバイダ (optional)>
-- }

local redis_client = redis:new()

local function get(key)
   logger.debug("session_manager.lua", "get " .. key)
   if (not key) then
      return nil, {message = "no key at session_manager.get"}
   end

   local val = redis_client:get("session:" .. key)
   if val then
      return json_safe.decode(val)
   end
end

local function set(key, session, expire)
   logger.debug("session_manager.lua", "set " .. key .. " expiring", expire)
   if (not key) or (not session) or (not expire) then
      return {message = "no key at session_manager.set"}
   elseif not redis_client:setex("session:" .. key, json_safe.encode(session), expire) then
      return {message = "session_manager.set " .. key .. " failed"}
   end
end

-- セッションを発行する。
local function generate_session(access_token, key_length, id_provider_uuid)
   key_length = key_length or 32
   local session = {}
   session.key = openssl_rsa.random_string(key_length)
   session.access_token = access_token.id
   session.expires = tonumber(os.date("%s")) + access_token.expires_in
   if id_provider_uuid then
      session.id_provider_uuid = id_provider_uuid
   end

   local err = set(session.key, session, access_token.expires_in)
   if err then
      return nil, {message = err}
   end

   return session
end

return {
   get = get,
   generate_session = generate_session
}
