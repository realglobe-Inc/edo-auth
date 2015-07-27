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


# ユーザー認証手続き代行バックエンド

ユーザー認証手続きを代行する。


## 1. 動作仕様

以降の動作記述において、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


### 1.1. エンドポイント

|エンドポイント名|初期 URI|機能|
|:--|:--|:--|
|ユーザー認証開始|/ 以下|認証機構にリダイレクトさせる|
|リダイレクト|/callback|ユーザー認証の後処理をして、アカウント情報を返す|


### 1.2. セッション

全エンドポイントでセッションを利用する。

|Cookie 名|値|
|:--|:--|
|Auth-User-Backend|セッション ID|

ユーザー認証開始エンドポイントへのリクエスト時に、セッションを発行し、レスポンス時にセッション ID を通知する。


### 1.3. ユーザー認証開始

認証に必要なパラメータを付加して、認証機構にリダイレクトさせる。

リクエストパスをセッションに紐付ける。

リクエストヘッダから以下のパラメータを取り出す。

|ヘッダ名|必要性|値|
|:--|:--|:--|
|X-Auth-Uri|必須|認証機構の URI。パラメータが付加されていても良い|

* `response_type` パラメータを `code` で追加または上書きする。
* 認証機構の URI が ID プロバイダの認証エンドポイントの場合、セッションに ID プロバイダを紐付ける。
* そうでない場合、`response_type` に `id_token` を追加する。
* `scope` パラメータに `openid` を追加する。
* `client_id`, `redirect_uri` パラメータが無ければ、既定値で追加する。
* `state`, `nonce` 値を生成し、セッションに紐付け、パラメータを追加または上書きする。


#### 1.3.1. レスポンス例

```http
HTTP/1.1 302 Found
Location: https://selector.example.org/?response_type=code%20id_token
    &scope=openid&client_id=https%3A%2F%2Fta.example.org
    &redirect_uri=https%3A%2F%2Fta.example.org%2Freturn&state=Ito-lCrO2H
    &nonce=v46QjbP6Qr
```


### 1.4. リダイレクト

ユーザー認証後のアカウント情報取得を行い、元のリクエストパスへのリダイレクトを行う。

* `state` がセッションに紐付くものと異なる、または、付加されたパラメータがエラーや不正な場合、
    * エラーを返す。
* そうでなければ、ID プロバイダからアクセストークンを取得する。

必要なら ID プロバイダからアカウント情報も取得する。

* `nonce` がセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなければ、アクセストークンを保存する。
  以下のセッションを発行する。
  以下のパラメータを付加して、セッションに紐付く元のリクエストパスにリダイレクトさせる。

|Cookie 名|値|
|:--|:--|
|Auth-User|セッション ID|

|HTTP ヘッダ名|値|
|:--|:--|
|X-Auth-User|クレームセットとしてアカウント情報を含む `alg` が `none` な [JWT]|

アカウント情報は以下の要素を含む。

* **`iss`**
    * アカウントが属す ID プロバイダ の ID。
* **`sub`**
    * アカウント ID。
* **`at_tag`**
    * edo-access-proxy を通した TA 間連携等で使うアクセストークンタグ。
* **`at_exp`**
    * アクセストークンの有効期限。
* その他の属性
    * 何が含まれるかは、ユーザー認証機構に何を要求するか、ユーザーが何を許可するかによる。


#### 1.4.1. リクエスト例

```http
GET /return?code=AFnKabazoCv99dVErDtxs5RYVmwh6R
    &id_token=eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIiwiY
    19oYXNoIjoibThIOGowbG5MZDZrN3FEZFNZVENqdyIsImV4cCI6MTQyNjU1ODI2MiwiaWF0IjoxN
    DI2NTU3NjYyLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsIm5vbmNlIjoidjQ2UWpiU
    DZRciIsInN1YiI6IjE5NTA0MTYyOTc3M0FFQ0MifQ.vevlIy6dviR6Khj8XX-zJttxEbSRych8PI
    wnCQpfTttMMok2xQJu0Pgg2y5a336NOZnQLgJZgLSN4QldZb-oFA&state=Ito-lCrO2H
Host: ta.example.org
Cookie: Auth-User-Backend=vmU7_v0qxDaCEg-8dHCNANAPVL-8Lj
```

改行とインデントは表示の都合による。


#### 1.4.2. レスポンス例

```http
HTTP/1.1 302 Found
Location: /ui/index.html
X-Auth-User: eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MjY1NjEyNjIsImF0X3RhZyI6IjJFeXd
    oMVo0dFoiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsInN1YiI6IjE5NTA0MTYyOTc
    3M0FFQ0MifQ.
```


### 1.5. エラーレスポンス

セッションを破棄し、ユーザーに理解し易いエラーページを返す。


<!-- 参照 -->
[JWT]: https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-32
