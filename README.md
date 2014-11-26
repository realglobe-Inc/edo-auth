edo-auth
=======


TA 認証
---

+ [nginx_proxy/lua/auth_ta.lua](nginx_proxy/lua/auth_ta.lua): Lua による TA 認証用 nginx モジュール。
+ [nginx_proxy/install_auth_ta.sh](nginx_proxy/install_auth_ta.sh): TA 認証を組み込んだ nginx をセットアップするスクリプト。
+ [nginx_proxy/test_auth_ta.sh](nginx_proxy/test_auth_ta.sh): TA 認証を組み込んだ nginx の動作チェック用スクリプト。
    + [nginx_proxy/sample/public_keys/auth-ta-checker-no-id.pub](nginx_proxy/sample/public_keys/auth-ta-checker-no-id.pub): 検証用公開鍵。
    + [nginx_proxy/sample/private_keys/auth-ta-checker-no-id.key](nginx_proxy/sample/private_keys/auth-ta-checker-no-id.key): 署名用秘密鍵。
+ [nginx_proxy/sample/nginx.auth_ta.conf](nginx_proxy/sample/nginx.auth_ta.conf): TA 認証を組み込んだ nginx.conf の例。
  install_auth_ta.sh のデフォルト設定。


### 設置
nginx_proxy/install_auth_ta.sh を実行すると、nginx_proxy/lib 以下に必要なプログラムがインストールされ、実行される。


### 設定
nginx_proxy/sample/nginx.auth_ta.conf を参照。
基本的に、nginx.conf の中で auth_ta.lua を読み込む前に、

    set $edo_auth_<パラメータ名> <値>;

の形で設定する。
パラメータは nginx_proxy/lua/auth_ta.lua の前の方に列挙してある。

### 動作
仕様は [edo/doc/spec_edo_auth_ta.md](https://github.com/realglobe-Inc/edo/blob/development/doc/spec_edo_auth_ta.md) を参照。
ただし、現状、TA 検証用公開鍵の取得はファイルからのみで、外部からは取得しない。

#### TA 検証用公開鍵の取得
現状、ローカルファイルか redis にキャッシュされている鍵のみ読める。
ローカルファイルは、nginx.conf で $edo_auth_public_key_directory に設定したディレクトリ内に &lt;TA の ID&gt;.pub のファイル名で置く。
中身は PEM 形式。

ローカルファイルをリモートと同期するような機能は無いので、必要ならば別口で行う。
例えば、

    wget -N -r -l 1 -P /var/edo/ta/public_keys -nH --cut-dirs=3 --no-parent -R 'index.html?*' https://example.org/edo/ta/public_keys/

のようなコマンドを cron で定期的に実行する。


アカウント認証
---
未定


---


add_header.lua, decrypt.lua, login.lua, callback.lua 等
---
仕様模索段階でのプロトタイプ。

### 設置方法

nginx_proxy/install.sh を参照。  
その後、nginx_proxy/lua/lib/config.lua と nginx.conf を適切に設定し nginx を起動する。  

### テスト

nginx_proxy/test.sh を参照。  

### SSO(OAuth2)モジュール

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


proxy.lua 等
---
location を分けずに動かすのを前提とした試作。
独自のアカウント認証に対応する。