package.path = package.path..";"..ngx.var.lua_lib_dir.."/?.lua"

local config = require "config"
local logger = require "logger"
local randomizer = require "randomizer"
local cookie_manager = require "cookie_manager"

local state = randomizer.generate(16)
local expires_at_unix_timestamp = tonumber(os.date("%s")) + config.oauth.state_expire_seconds
local expires = ngx.cookie_time(expires_at_unix_timestamp)

logger.debug("login.lua", "expires:", expires)
cookie_manager.set("oauth_state", state, {path = "/", expires = expires})
logger.debug("login.lua", "Set-Cookie header", ngx.header["Set-Cookie"])

local request_params = {
   client_id = config.oauth.client_id,
   redirect_uri = config.oauth.callback_url,
   response_type = "code",
   scope = config.oauth.scope,
   state = state
}
local params_string = ngx.encode_args(request_params)
local url = config.oauth.authorize_url.."?"..params_string
ngx.redirect(url)
