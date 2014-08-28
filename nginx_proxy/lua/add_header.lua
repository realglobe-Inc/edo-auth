package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local config = require "config"
local logger = require "logger"
local exiter = require "exiter"
local token_manager = require "token_manager"

local session_key = ngx.var.cookie_oauth_session_key
local access_token
local id_token

logger.debug("add_header.lua", "session_key:", session_key)
if session_key and session_key ~= "" then
   access_token = token_manager.access_token.get(session_key)
   id_token = token_manager.id_token.get(session_key)
end

ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", access_token)
ngx.req.set_header("X-OAUTH-ID-TOKEN", id_token)

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
