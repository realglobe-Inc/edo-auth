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


local cjson = require("cjson.safe")


local function respond_html(params)
   local status = ngx.HTTP_INTERNAL_SERVER_ERROR
   local message = "error occurred"
   local others = ""
   if params.status then
      status = params.status
      params.status = nil
   end
   if params.message then
      message = params.message
      params.message = nil
   end
   for k, v in pairs(params) do
      others = others .. [[<b>]] .. k .. [[:</b> ]] .. v .. [[<br/>]]
   end

   local body = [[
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>Error</title>
  </head>

  <body>
    <h1>]]
      .. status ..
      [[</h1>
    <p>
      <b>error:</b> ]]
      .. message ..
      [[<br/>]]
      .. others ..
      [[</p>
  </body>
</html>]];

   -- レスポンスの中身を準備できた。
   ngx.log(ngx.ERR, "respond error: " .. status .. " " .. message)

   ngx.header["Content-Type"] = "text/html"
   ngx.status = status
   ngx.print(body)
   return ngx.exit(status)
end


local function respond_json(params)
   local status = params.status or ngx.HTTP_INTERNAL_SERVER_ERROR
   local message = params.message or "error occurred"

   local body, err = cjson.encode(params)
   if err then
      ngx.log(ngx.ERR, err)
      body = '{"status"=' .. status .. '}'
   end

   -- レスポンスの中身を準備できた。
   ngx.log(ngx.ERR, "respond error: " .. status .. " " .. message)

   ngx.header["Content-Type"] = "application/json"
   ngx.status = status
   ngx.print(body)
   return ngx.exit(status)
end


return {
   respond_html = respond_html,
   respond_json = respond_json,
}
