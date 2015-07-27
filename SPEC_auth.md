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


# ユーザー認証手続き代行フロントエンド

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
|ユーザー認証代行|/ 以下|アカウント情報を付加する|
|リダイレクト|/callback|アカウント情報を保存する|
|更新|/refresh|アクセストークンを更新する|


### 1.2. セッション

全エンドポイントでセッションを利用する。

|Cookie 名|値|
|:--|:--|
|Auth-User|セッション ID|

リダイレクトエンドポイントでの成功時に、セッションを登録する。


### 1.3. ユーザー認証代行

リクエストにアカウント情報を付加する。

* アカウント情報と紐付くセッションの場合、
    * Cookie から Auth-User を除去する。
      セッションに紐付くアカウント情報を付加する。
* そうでなければ、X-Auth-Authentication-Uri ヘッダを加えて、バックエンドの認証開始エンドポイントに渡す。


#### 1.3.1. アカウント情報の付加

HTTP ヘッダにて行う。

|HTTP ヘッダ名|値|
|:--|:--|
|X-Auth-User|アカウント情報。詳細はバックエンドを参照のこと|


##### 1.3.1.1. 例

元のリクエストは、

```http
GET /ui/index.html HTTP/1.1
Host: ta.example.org
Cookie: Auth-User-Frontend=vmU7_v0qxDaCEg-8dHCNANAPVL-8Lj
```

通過後は、

```http
GET /ui/index.html HTTP/1.1
Host: ta.example.org
X-Auth-User: eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MjY1NjEyNjIsImF0X3RhZyI6IjJFeXd
    oMVo0dFoiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsInN1YiI6IjE5NTA0MTYyOTc
    3M0FFQ0MifQ.
```

改行とインデントは表示の都合による。

クレームセットの内容は、

```json
{
    "iss": "https://idp.example.org",
    "sub": "195041629773AECC",
    "at_tag": "2Eywh1Z4tZ",
    "at_exp": 1426561262
}
```

改行とインデントは表示の都合による。


### 1.4. リダイレクトエンドポイント

ユーザー認証機構へのリダイレクト時に `redirect_uri` パラメータとして与えるエンドポイント。
ユーザー認証後のアカウント情報の保存を行う。

* バックエンドのリダイレクトエンドポイントにリクエストを転送する。
* バックエンドからのレスポンスの X-Auth-User ヘッダの値を取り出す。
  取り出した値をアカウント情報としてセッションに紐付ける。
  レスポンスから X-Auth-User ヘッダを削除し、ユーザーに返す。


### 1.5. 更新エンドポイント

edo-access-proxy を通した TA 間連携等のためにアクセストークンを意図的に取得する。

リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|`path`|必須|リクエストパス|

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなければ、`path` をリクエストパスにして、X-Auth-Authentication-Uri ヘッダを加えて、バックエンドの認証開始エンドポイントに渡す。


### 1.6. エラーレスポンス

ユーザーに理解し易いページを返す。


<!-- 参照 -->
