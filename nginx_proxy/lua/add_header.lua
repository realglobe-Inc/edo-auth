package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local json_safe = require "cjson.safe"
local config = require "config"
local logger = require "logger"
local redis = require "redis"
local exiter = require "exiter"
local curl_wrapper = require "curl_wrapper"

local session_key = ngx.var.cookie_oauth_session_key
local access_token

logger.debug("add_header.lua", "session_key:", session_key)
logger.debug("add_header.lua", "redis", redis)

if session_key and session_key ~= "" then
   local redis_client = redis:new()
   logger.debug("add_header.lua", "redis_client", redis_client)
   access_token = redis_client:get(session_key)
end

-- TODO
-- if access_token then
if false then
   -- TODO
   local user_api_endpoint = "https://api.github.com/user"
   local headers = {
      "User-Agent: curl/"..cURL.version_info().version,
      "Authorization: token "..access_token,
   }
   local curl_client = curl_wrapper:new()
   curl_client:set_headers(headers)
   local response = curl_client:get(user_api_endpoint)
   logger.debug("add_header.lua", "response.body", response.body)
   if response.body then
      local response_object = json_safe.decode(response.body)
      -- TODO
      ngx.req.set_header("X-OAUTH-ID", response_object["id"])
   end
end

ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", access_token)

if config.oauth.force_login and access_token == nil then
   logger.debug("add_header.lua", "request_url:", ngx.var.uri)
   logger.debug("add_header.lua", "redirect_to login_path:", config.oauth.login_path)
   if ngx.var.uri == config.oauth.login_path then
      -- TODO error message
      exiter.exit("request path is login_endpoint")
   else
      ngx.redirect(config.oauth.login_path)
   end
end
