oauth_module_path = os.getenv("OLDPWD")

return {
   service_uuid = "service_uuid",
   redis = {
      unix_socket_path = "unix:"..oauth_module_path.."/lib/redis/bin/redis.sock",
      connection_timeout_millisec = 1000
   },
   public_key = {
      local_storage_directory = oauth_module_path.."/tmp/public_keys/",
      api_endpoint = "file://"..oauth_module_path.."/mock/public_key"
   },
   cookie = {
      secure = true,
      http_only = true
   },
   oauth = {
--      force_login = true,
      client_id = "test_id",
      client_secret = "test_secret",
      authorize_url = "https://localhost/login/oauth/authorize",
      callback_url = "https://localhost/oauth/callback",
      scope = "user",
      state_expire_seconds = 300,
      access_token_endpoint = "file://"..oauth_module_path.."/mock/access_token.json",
      access_token_default_expires_in = 3600,
      login_path = "/oauth/login",
      after_login_path = "/index.html",
   },
   rsa = {
      token_lifetime = 600
   }
}
