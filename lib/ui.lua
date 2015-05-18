-- Copyright 2015 realglobe, Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.


local function respond_error(params)
   local body = [[<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Error</title>
</head>
<body>
<h1>Error</h1>
<p>]];

   local status = ngx.HTTP_INTERNAL_SERVER_ERROR
   local log_msg = ""
   if params.status then
      body = body .. [[<b>status:</b> ]] .. params.status .. [[<br/>]]
      status = params.status
      log_msg = log_msg .. params.status
      params.status = nil
   end
   if params.message then
      body = body .. [[<b>message:</b> ]] .. params.message .. [[<br/>]]
      if #log_msg > 0 then
         log_msg = log_msg .. ": "
      end
      log_msg = log_msg .. params.message
      params.message = nil
   end
   if params.path then
      body = body .. [[<b>path:</b> ]] .. params.path .. [[<br/>]]
      params.path = nil
   end
   for k, v in pairs(params) do
      body = body .. [[<b>]] .. k .. [[:</b> ]] .. v .. [[<br/>]]
   end

   body = body .. [[</p>
</body>
</html>]];

   -- レスポンスの中身を準備できた。
   ngx.log(ngx.ERR, "respond error " .. log_msg)

   ngx.header["Content-Type"] = "text/html"
   ngx.header["Content-Length"] = string.len(body) .. ""
   ngx.status = status
   ngx.print(body)
   return ngx.exit(status)
end

return {
   respond_error = respond_error,
}
