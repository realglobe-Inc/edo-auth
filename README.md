<!--
   Copyright 2015 realglobe, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
-->


# edo-auth

現状は、TA 内にて、アクセスしてきた他の TA の認証を行う nginx のモジュール。

以下、例における /opt/edo-auth はインストールパスとする。
異なる場合は適宜置き換えること。


## 1. インストール

install.sh で nginx ごとセットアップできる。

```shell
./install.sh
```

インストール先を変える場合は install_dir で指定する。

```shell
install_dir=/opt/edo-auth ./install.sh
```


### 1.1. インストールオプション

|シェル変数|初期値|値|
|:--|:--|:--|
|install_dir|install.sh のあるディレクトリ|インストール先のパス|
|nginx_port|7000|nginx のポート番号|
|nginx|true|インストール終了時に nginx を起動するかどうか|
|proxy_pass|http://example.org|edo-auth を通した後に転送する先|


### 1.2. 既存の nginx で使う

必要なモジュールと共に nginx がコンパイルされているなら、nginx の設定ファイルの適切な箇所に以下のような設定を追加する。

```nginx
set $edo_auth_dir /opt/edo-auth;
set $edo_auth_public_key_directory $edo_auth_dir/public_keys;
access_by_lua_file $edo_auth_dir/lua/auth_ta.lua;
```

## 2. 実行


### 2.1. DB の準備

動作には [redis](http://redis.io/) が必要となる。
インストールされていない場合はインストールする。


### 2.2. 公開鍵ファイルの設置

許可する TA の公開鍵を、&lt;TA の ID&gt;.pub という名前で公開鍵ディレクトリに置く。

```
<公開鍵ディレクトリ>/
├── <TA 1 の ID>.pub
├── <TA 2 の ID>.pub
...
```

公開鍵ディレクトリのパスは設定ファイルで指定する。


### 2.3. 起動

install.sh でインストールした場合はデーモンとして起動している。
起動させていない場合、一旦終了させた場合は、

```shell
/opt/edo-auth/nginx/sbin/nginx
```


### 2.4. 設定

設定は nginx の設定ファイルの中で auth\_ta.lua を読む前に書く。

|項目名|初期値|値|
|:--|:--|:--|
|$edo\_auth\_log\_level|debug|edo-auth のログを nginx のどのログレベルで出力するか|
|$edo\_auth\_redis\_host|127.0.0.1|使用する redis のアドレス|
|$edo\_auth\_redis\_port|6379|使用する redis のポート番号|
|$edo\_auth\_public\_key\_directory||公開鍵置き場のパス|


## 3. 動作仕様

リクエストが edo-auth を通ったら、通信元 TA の認証が済んでいることを保証する。


### 3.1. 概要

認証済みの TA からのリクエストなら X-Edo-Ta-Id ヘッダでその TA の ID を付加し、未認証ならば認証を行う。


### 3.2. リクエストの受け取り

+ Cookie に X-Edo-Ta-Seesion が無ければ、未認証 TA からのリクエストとして処理する。
+ X-Edo-Auth-Ta-Session の値が認証中セッションであれば、認証中 TA からのリクエストとして処理する。
+ X-Edo-Auth-Ta-Session の値が認証済みセッションであれば、認証済み TA からのリクエストとして処理する。
+ それ以外なら、未認証 TA からのリクエストとして処理する。


### 3.3. 認証済み TA からのリクエストの処理

リクエストから Cookie の X-Edo-Auth-Ta-Session を取り除き、以下のヘッダを加えて通す。

|ヘッダ名|値|
|:--|:--|
|X-Edo-Ta-Id|セッションと紐付く TA の ID|


### 3.4. 認証中 TA からのリクエストの処理

認証中 TA からのリクエストは以下の HTTP ヘッダを含まなければならない。

|ヘッダ値|値|
|:--|:--|
|X-Edo-Auth-Ta-Id|通信元 TA の ID|
|X-Edo-Auth-Ta-Token-Sign|セッションと紐付く乱数列への署名|

また、追加で以下のヘッダを含むこともある。

|ヘッダ値|値|
|:--|:--|
|X-Edo-Auth-Hash-Function|X-Edo-Auth-Ta-Token-Sign の署名に使ったハッシュ関数|

通信元 TA の公開鍵 (と X-Edo-Auth-Hash-Function で指定されたハッシュ関数) で、X-Edo-Auth-Ta-Token-Sign の値を検証する。
問題が無ければ、セッションと通信元 TA を紐付け、認証済み TA として処理する。


### 3.5. 未認証 TA からのリクエストの処理

ランダムな文字列を生成し、セッションを作成し、セッションと生成した文字列を紐付け、以下の Set-Cookie とヘッダを含む 401 Unauthorized レスポンスを返す。

|Set-Cookieラベル|値|
|:--|:--|
|X-Edo-Auth-Ta-Session|セッション|

|ヘッダ名|値|
|:--|:--|
|X-Edo-Auth-Ta-Token|生成した文字列|
|X-Edo-Auth-Ta-Error|"start new session"|


### 3.6. エラーレスポンス

edo-auth にてエラーが発生した場合、以下のヘッダを加えたレスポンスを返す。

|HTTP ヘッダフィールド|値|
|:--|:--|
|X-Edo-Auth-Ta-Error|適当なメッセージ|

+ セッションの通信元アドレスが異なる場合、403 Forbidden を返す。
+ 認証中に X-Edo-Auth-Ta-Id ヘッダおよび X-Edo-Auth-Ta-Token-Sign ヘッダが無い場合、403 Forbidden を返す。
+ 通信元 TA の公開鍵が登録されていない場合、403 Forbidden を返す。
+ X-Edo-Auth-Ta-Token-Sign の検証に失敗した場合、403 Forbidden を返す。


## 4. ライセンス

Apache License, Version 2.0
