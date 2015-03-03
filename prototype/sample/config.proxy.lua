-- proxy.lua 用の config.lua の例。

return {
   service_uuid = "edo-test-ui-with-nginx-no-uuid",
   redis = {
      unix_socket_path = "unix:/home/mui/work.git/src/github.com/realglobe-Inc/edo-auth/nginx_proxy/lib/redis/bin/redis.sock",
--      connection_timeout_millisec = 1000
   },
   public_key = {
      local_storage_directory = "/home/mui/work.git/src/github.com/realglobe-Inc/edo-auth/nginx_proxy/tmp/public_keys",
      api_endpoint = "http://localhost:9002/" -- service_key_registry
   },
   cookie = {
      secure = false, -- デバッグのため。
      http_only = true
   },
   oauth = {
      force_login = true,
      -- client_id は service_uuid を使う。
      -- client_secret は秘密鍵と code か access_token から生成する。
      authorize_url = "http://localhost:8000/", -- edo-directory-service
      -- callback_url は元のリクエスト URL。
      -- scope は edo-id-provider が未対応。
      state_expire_seconds = 300,
      access_token_default_expires_in = 3600,
      -- login_path, after_login_path は使わない。
      id_provider_registry_uri = "http://localhost:9001/", -- id-provider-attribute-registry
      private_key_directory = "/home/mui/work.git/src/github.com/realglobe-Inc/edo-auth/nginx_proxy/sample/private_keys"
   },
   rsa = {
      token_lifetime = 600
   }
}
