#!/bin/bash

cat <<EOF
return {
   service_uuid = "service-uuid", -- TODO
   redis = {
      unix_socket_path = "unix:`pwd`/lib/redis/bin/redis.sock",
--      connection_timeout_millisec = 1000
   },
   public_key = {
      local_storage_directory = "`pwd`/tmp/public_keys",
      api_endpoint = "https://public-key.server.com/api/public_keys/get" -- TODO
   },
   cookie = {
      secure = true,
      http_only = true
   },
   oauth = {
      force_login = true,
      client_id = "client_id", -- TODO
      client_secret = "client_secret", -- TODO
      authorize_url = "https://auth.server.com/oauth/authorize", -- TODO
      callback_url = "https://service.com/oauth/callback", -- TODO
      scope = "user",
      state_expire_seconds = 300,
      access_token_endpoint = "https://auth.server.com/oauth/access_token", -- TODO
      access_token_default_expires_in = 3600,
      login_path = "/oauth/login",
      after_login_path = "/",
   },
   rsa = {
      token_lifetime = 600
   }
}
EOF
