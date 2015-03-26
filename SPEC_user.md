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


# edo-auth user の仕様（目標）

以降の動作記述において、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


## 1. エンドポイント

|エンドポイント名|初期 URI|機能|
|:--|:--|:--|
|ユーザー認証処理代行|/ui 以下|ユーザー情報を付加して通す|
|リダイレクト|/return|ユーザー認証の後処理|
|更新|/refresh|アクセストークンを更新する|


## 2. セッション

セッションを利用する。

|Cookie ラベル|値|
|:--|:--|
|X-Edo-Auth-User|セッション ID|

ユーザー認証処理代行へのリクエスト時に、セッション ID が通知されなかった、または、期限切れだった場合、セッションを発行する。

ユーザー認証処理代行からユーザー認証機構へのリダイレクト時、また、リダイレクトエンドポイントから元のリクエスト先へのリダイレクト時に、未通知のセッション ID を通知する。


## 3. ユーザー認証処理代行

ユーザー認証処理を代行し、結果のユーザー情報をリクエストに付加する。

* ユーザー情報と紐付くセッションの場合、
    * Cookie からセッション情報を除去する。
      セッションに紐付くユーザー情報を付加して通す。
* そうでなければ、リクエスト内容をセッションに紐付ける。
  ユーザー認証機構（IdP または IdP 選択サービス）にリダイレクトさせる。



### 3.1. ユーザー情報の付加

HTTP ヘッダにて行う。

|HTTP ヘッダ名|値|
|:--|:--|
|X-Edo-User|クレームセットとしてユーザー情報を含む `alg` が `none` な [JWT]|

ユーザー情報は以下の要素を含む。

* **`iss`**
    * IdP の ID。
* **`sub`**
    * アカウント ID。
* **`at_tag`**
    * edo-access-proxy を通した TA 間連携等で使うアクセストークンタグ。
* **`at_exp`**
    * アクセストークンの有効期限。

その他にどんな要素が含まれるかは、ユーザー認証機構に何を要求するか、ユーザーが何を許可するかによる。


#### 3.1.1. 例

元のリクエストは、

```http
GET / HTTP/1.1
Host: ta.example.org
Cookie: X-Edo-Auth-User=vmU7_v0qxDaCEg-8dHCNANAPVL-8Lj
```

通過後は、

```http
GET / HTTP/1.1
Host: ta.example.org
X-Edo-User: eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MjY1NjEyNjIsImF0X3RhZyI6IjJFeXdo
    MVo0dFoiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsInN1YiI6IjE5NTA0MTYyOTc3
    M0FFQ0MifQ.
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


### 3.2. ユーザー認証機構へのリダイレクト

`state` と `nonce` を発行する。
`state` と `nonce` をセッションに紐付ける。
認証パラメータを付加したユーザー認証機構のエンドポイントにリダイレクトさせる。


#### 3.2.1. レスポンス例

```http
HTTP/1.1 302 Found
Set-Cookie: X-Edo-Auth-User=vmU7_v0qxDaCEg-8dHCNANAPVL-8Lj;
    Expires=Tue, 24 Mar 2015 01:59:18 GMT; Path=/; Secure; HttpOnly
Location: https://selector.example.org/?response_type=code%20id_token
    &scope=openid&client_id=https%3A%2F%2Fta.example.org
    &redirect_uri=https%3A%2F%2Fta.example.org%2Freturn&state=Ito-lCrO2H
    &nonce=v46QjbP6Qr
```

改行とインデントは表示の都合による。


## 4. リダイレクトエンドポイント

ユーザー認証機構へのリダイレクト時に `redirect_uri` パラメータとして与えるエンドポイント。
ユーザー認証後のセッション更新とリダイレクトを行う。

* `state` がセッションに紐付くものと異なる、または、付加されたパラメータがエラーや不正な場合、
    * エラーを返す。
* そうでなければ、IdP からアクセストークンを取得する。

必要なら IdP からユーザー情報も取得する。

* `nonce` がセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなければ、設定を引き継いだセッションを発行する。
  アクセストークンとユーザー情報をセッションに紐付ける。
  セッションに紐付くリクエスト内容を再現するようにリダイレクトさせる。


### 4.1. リクエスト例

```http
GET /return?code=AFnKabazoCv99dVErDtxs5RYVmwh6R
    &id_token=eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIiwiY
    19oYXNoIjoibThIOGowbG5MZDZrN3FEZFNZVENqdyIsImV4cCI6MTQyNjU1ODI2MiwiaWF0IjoxN
    DI2NTU3NjYyLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsIm5vbmNlIjoidjQ2UWpiU
    DZRciIsInN1YiI6IjE5NTA0MTYyOTc3M0FFQ0MifQ.vevlIy6dviR6Khj8XX-zJttxEbSRych8PI
    wnCQpfTttMMok2xQJu0Pgg2y5a336NOZnQLgJZgLSN4QldZb-oFA&state=Ito-lCrO2H
Host: ta.example.org
Cookie: X-Edo-Auth-User=vmU7_v0qxDaCEg-8dHCNANAPVL-8Lj
```

改行とインデントは表示の都合による。


### 4.2. レスポンス例

```http
HTTP/1.1 302 Found
Set-Cookie: X-Edo-Auth-User=UpmP-WGyxducqFqEviJyQnVjPdpZ1Q;
    Expires=Tue, 24 Mar 2015 02:01:10 GMT; Path=/; Secure; HttpOnly
Location: /
```


## 5. 更新エンドポイント

edo-access-proxy を通した TA 間連携等のためにアクセストークンを意図的に取得する。

リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|`redirect_uri`|必須|リクエスト内容|

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなければ、`redirect_uri` をリクエスト内容としてセッションに紐付ける。
  ユーザー認証機構にリダイレクトさせる。


## 6. 外部データ

以下に分ける。

* 共有データ
    * 他のプログラムと共有する可能性のあるもの。
* 非共有データ
    * 共有するとしてもこのプログラムの別プロセスのみのもの。


### 6.1. 共有データ


#### 6.1.1. IdP 情報

以下を含む。

* ID
* 検証鍵
* トークンエンドポイント
* ユーザー情報エンドポイント

以下の操作が必要。

* ID による取得


#### 6.1.2. アクセストークン

edo-access-proxy の TA 間連携等で利用する。

以下を含む。

* ID
* タグ
* 有効期限

以下の操作が必要。

* 保存
* タグによる取得


### 6.2. 非共有データ


#### 6.2.1. セッション

以下を含む。

* ID
* アクセストークン
* ユーザー情報
* リクエスト内容
* チケット

以下の操作が必要。

* 保存
* ID による取得
* 上書き


<!-- 参照 -->
[JWT]: https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-32