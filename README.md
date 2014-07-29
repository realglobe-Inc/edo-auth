edo-auth
=======

### SSO(OAuth2)モジュール

#### 設置方法

まずアプリをOAuthクライアントとして登録し、クライアントIDやシークレットトークンを取得しておく。  
nginx の起動時に、nginx.conf で以下の項目を適切に設定する。  

* 認証サーバーのURL、クライアントとして登録してあるリダイレクトURL、クライアントID、シークレットトークンなど  
* ログインURL(login.lua にルーティング)、リダイレクトURL(callback.lua にルーティング)  
* OAuthサーバーのアクセストークンエンドポイント  
* redis のURL or unixソケット  
* バックエンドのアプリにプロキシする前に、add_header.lua を通す  

login.lua が実行されると認証サーバーにリダイレクトされ、認証されると callback.lua にリダイレクトされる。  
callback.lua ではアクセストークンを取得し、redis に保存する。アクセストークンが正常に取得できれば cookie: oauth_session_key が設定され、ログイン状態になる。  
ログイン状態なら X-OAUTH-ACCESS-TOKEN ヘッダにアクセストークンが設定されるので、必要に応じてアプリから使用する。  

### RSA認証モジュール

メッセージの送信者は、以下の二通りのうちどちらかの方法でリクエストヘッダを付与する。  

#### 生のトークンを署名する場合

* X-EDO-Private-Key-UUID
  * 送信者が暗号化に使用した秘密鍵のUUID。送信者のUUIDとは異なる場合もある
* X-EDO-Auth-Encoded-Token
  * "送信者のUUID,受信者のUUID,UNIXタイムスタンプ"を「base64エンコード+改行を削除」したもの
* X-EDO-Auth-Params
  * オプション。任意のJSONを「base64エンコード+改行を削除」したもの

```sh
$ sender_uuid=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
$ receiver_uuid=yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
$ private_key_uuid=zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz
$ timestamp=`date +%s`
$ encoded_token=`printf ${sender_uuid},${receiver_uuid},${timestamp} | openssl rsautl -sign -inkey /path/to/private.key | openssl base64 | tr -d '\n'`

$ curl https://edo-service.com/ -H "X-EDO-Private-Key-UUID: ${private_key_uuid}" -H "X-EDO-Auth-Encoded-Token: ${encoded_token}" -H "X-EDO-Auth-Params: ..."
```

#### ハッシュ値を署名する場合

* X-EDO-Private-Key-UUID
  * 送信者が署名に使用した秘密鍵のUUID。送信者のUUIDとは異なる場合もある
* X-EDO-Auth-Token
  * "送信者のUUID,受信者のUUID,タイムスタンプ"
* X-EDO-Auth-Signed-Token
  * 送信者が作成した署名を「base64エンコード+改行を削除」したもの
* X-EDO-Hash-Function
  * 送信者が署名に使用したハッシュ関数。md5、sha1、sha256、など
* X-EDO-Auth-Params
  * オプション。任意のJSONを「base64エンコード+改行を削除」したもの

```sh
$ sender_uuid=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
$ receiver_uuid=yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
$ private_key_uuid=zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz
$ timestamp=`date +%s`
$ auth_token=${sender_uuid},${receiver_uuid},${timestamp}
$ signed_token=`printf ${auth_token} | openssl dgst -sign /path/to/private.key -sha256 | openssl base64 | tr -d '\n'`

$ curl https://edo-service.com/ -H "X-EDO-Private-Key-UUID: ${private_key_uuid}" -H "X-EDO-Auth-Token: ${auth_token}" -H "X-EDO-Auth-Signed-Token: ${signed_token}" -H "X-EDO-Hash-Function: sha256" -H "X-EDO-Auth-Params: ..."
```
