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


-- 公開鍵 DB。

-- ファイルを読み込む。
local function read_file(path)
   local fd, err = io.open(path)
   if err then
      -- ファイルが無かった？
      return nil, err
   end

   local buff = fd:read("*a")
   fd:close()
   if (not buff) or buff == "" then
      -- 空だった。
      return nil, path .. " is empty"
   end

   return buff
end

-- PEM または DER 形式の公開鍵を読んで、PEM 形式で返す。
local function read_pub(path)
   local buff, err = read_file(path)
   if err then
      return nil, err
   end

   -- 読めた。
   local public_key, err = openssl.pkey.read(buff)
   if err then
      -- ファイルの中身がおかしかった。
      return nil, err
   end
   return public_key:export()
end

-- PEM または DER 形式の証明書を読んで、公開鍵をPEM 形式で返す。
local function read_crt(path)
   local buff, err = read_file(path)
   if err then
      return nil, err
   end

   -- 読めた。
   local cert, err = openssl.x509.read(buff)
   if err then
      -- ファイルの中身がおかしかった。
      return nil, err
   end

   local public_key, err = cert:pubkey()
   if err then
      return nil, err
   end

   return public_key:export()
end

-- メソッド定義。
local db_redis = {
   -- 公開鍵を取得する。
   get = function(self, ta_id)
      -- redis を調べてから、ファイルを調べる。

      local _, err = self.redis:connect()
      if err then
         return err
      end

      local public_key_pem = self.redis.base:get(self.prefix .. ta_id)
      self.redis:close()
      if public_key_pem and public_key_pem ~= ngx.null then
         ngx.log(self.log_level, "public key is exist in redis")
         -- 公開鍵の更新時に再起動が必須になってしまうので、キャッシュの期限は更新しない方が良さそう。
         return public_key_pem
      end

      -- redis になかった。

      -- .pub または .crt から公開鍵を読んで、キャッシュする。
      local file_name_base = ngx.escape_uri(ta_id)
      local public_key_pem, err = read_pub(self.path .. "/" .. file_name_base .. ".pub")
      if err then
         -- 公開鍵ファイルに不具合。
         ngx.log(ngx.ERR, err)
      end

      if (not public_key_pem) or public_key_pem == ""  then
         public_key_pem, err = read_crt(self.path .. "/" .. file_name_base .. ".crt")
         if err then
            -- 証明書ファイルに不具合。
            ngx.log(ngx.ERR, err)
            return nil
         elseif (not public_key_pem) or public_key_pem == ""  then
            -- 公開鍵が無かった。
            return nil
         end
      end

      -- 公開鍵が読めた。
      ngx.log(self.log_level, "public key of " .. ta_id .. " is exist")

      -- キャッシュする。
      local _, err = self.redis:connect()
      if err then
         return err
      end

      local _, err = self.redis.base:set(self.prefix .. ta_id, public_key_pem, "EX", self.expires_in)
      self.redis:close()
      if err then
         ngx.log(ngx.ERR, err)
      else
         ngx.log(self.log_level, "public key was cached into redis")
      end

      return public_key_pem
   end,
}


-- redis ドライバとキーの接頭辞を指定して作成する。
new_redis = function(redis, prefix, path, expires_in, log_level)
   local obj = {
      redis = redis,
      prefix = prefix,
      path = path,
      expires_in = expires_in,
      log_level = log_level or ngx.DEBUG,
   }
   setmetatable(obj, {__index = db_redis})
   return obj
end


return {
   new_redis = new_redis,
}
