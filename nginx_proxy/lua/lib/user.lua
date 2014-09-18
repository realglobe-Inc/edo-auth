-- ユーザー認証周り。

local logger = require("logger")
local cookie_manager = require("cookie_manager")

-- ユーザー認証済みであることをヘッダに示す。
local function set_header(session)
   -- サービス認証済みと混同することのないように。
   ngx.req.clear_header("X-EDO-Sender-UUID")
   ngx.req.clear_header("X-EDO-Receiver-UUID")
   ngx.req.clear_header("X-EDO-Timestamp")

   ngx.req.set_header("X-OAUTH-ACCESS-TOKEN", session.access_token)
   if session.id_provider_uuid then
      ngx.req.set_header("X-OAUTH-ID-PROVIDER", session.id_provider_uuid)
   end
end

-- セッションを設定するために元のページにリダイレクト。
local function start_session(session)
   local redirect_uri = ngx.var.uri -- 相対パスで良い。

   -- code と id_provider_uuid を取り除く。
   local queries = {}
   for key, val in pairs(ngx.req.get_uri_args()) do
      if key ~= "code" and key ~= "id_provider_uuid" then
         queries.key = val
      end
   end
   local query_string = ngx.encode_args(queries)
   if query_string and query_string ~= "" then
      redirect_uri = redirect_uri .. "?" .. query_string
   end

   local expires = ngx.cookie_time(session.expires)
   cookie_manager.set("oauth_session_key", session.key, {path = "/", expires = expires})

   return ngx.redirect(redirect_uri)
end

-- ユーザー認証に旅立たせる。
local function authorize(client_id, authorize_uri)
   local query_redirect_uri = ngx.var.scheme .. "://" .. ngx.var.http_host .. ngx.var.request_uri
   local queries = ngx.encode_args({ client_id = client_id, redirect_uri = query_redirect_uri })
   return ngx.redirect(authorize_uri .. "?" .. queries)
end

return {
   set_header = set_header,
   start_session = start_session,
   authorize = authorize
}
