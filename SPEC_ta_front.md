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


# edo-auth TA フロントエンドの仕様（目標）

以降の動作記述において、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


## 1. 概要

1. 仲介コードと共にリクエストを受け取る。
2. 仲介コードを edo-auth バックエンドに渡す。
3. edo-auth バックエンドから仲介情報を受け取る。
4. リクエストに仲介情報を付加してバックエンドに通す。
5. バックエンドからレスポンスを受け取る。
6. レスポンスを転送する。

```
+--------+                           +--------+
|        |----(1) request + codes--->|        |                    +--------+
|        |                           |        |----(2) codes------>|edo auth|
|        |                           |        |                    |TA      |
|        |                           |        |                    |backend |
|        |                           |        |----(3) userinfo--->|        |
|        |                           |        |                    +--------+
|        |                           |edo auth|
|   TA   |                           |TA      |
|        |                           |frontend|
|        |                           |        |                    +--------+
|        |                           |        |----(4) request---->|        |
|        |                           |        |        +usrinfo    |backend |
|        |                           |        |                    |        |
|        |                           |        |<---(5) response----|        |
|        |<---(6) response-----------|        |                    +--------+
+--------+                           +--------+
```


## 2. リクエストの受け取り

[TA 間連携プロトコル]も参照のこと。

* Cookie の Edo-Cooperation によりセッションが宣言されている場合、
    * セッションが存在する場合、
        * リクエストの転送へ。
    * そうでなければ、
        * エラーを返す。
* そうでなく、仲介コードが付加されている場合、
    * edo-auth バックエンドへ仲介コードを送る。
* そうでなければ、エラーを返す。


## 3. edo-auth バックエンドへの仲介コードの送信

受け取った形式のまま送る。


## 4. edo-auth バックエンドからの仲介情報の受信

edo-auth バックエンドを参照のこと。

セッションが通知されたら、セッションと仲介情報を紐付けて保存する。


## 5. リクエストの転送

リクエストから仲介コードやセッション宣言を削除する。
以下の HTTP ヘッダにて仲介情報を付加する。

* X-Auth-User
    * 処理の主体の情報をクレームセットに含む `alg` が `none` な [JWT]。
* X-Auth-User-Tag
    * 処理の主体のアカウントタグ。
* X-Auth-Related-Users
    * 処理の主体でないアカウントが含まれる場合のみ。
      処理の主体でないアカウント全てについて、アカウントタグからアカウント情報へのマップをクレームセットに含む `alg` が `none` な [JWT]。
* X-Auth-From-Id
    * 要請元 TA の ID。

アカウント情報は以下の要素を含む。

* **`iss`**
    * IdP の ID。
* **`sub`**
    * アカウント ID。
* **`at_tag`**
    * 処理の主体のみ。アクセストークンタグ。
* **`at_exp`**
    * 処理の主体のみ。アクセストークンの有効期限。

その他にどんな要素が含まれるかは、IdP に何を要求するか、ユーザーが何を許可しているかによる。


### 5.1. 仲介情報の付加例

元のリクエストは、

```http
GET /api/writer/profile HTTP/1.1
Host: to.example.org
X-Edo-Cooperation-Codes: ...
```

仲介コードは省いている。

付加後は、

```http
GET /api/writer/profile HTTP/1.1
Host: to.example.org
X-Auth-User: eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MjY1NjEyNjIsImF0X3RhZyI6InVudG5G
    ZHhOMDMiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsInN1YiI6Ijc1NUI2MjkyMDhF
    REZEQzIifQ.
X-Auth-User-Tag: reader
X-Auth-From-Id: https://from.example.org
X-Auth-Related-Users: eyJhbGciOiJub25lIn0.eyJ3cml0ZXIiOnsiaXNzIjoiaHR0cHM6Ly9pZHA
    uZXhhbXBsZS5vcmciLCJzdWIiOiJENUNGM0Y0OUU1RTczMUMzIn19.
```

JWT の改行とインデントは表示の都合による。

X-Auth-User のクレームセットは、

```json
{
    "iss": "https://idp.example.org",
    "sub": "755B629208EDFDC2",
    "at_tag": "untnFdxN03",
    "at_exp": 1426561262
}
```

X-Auth-Related-Users のクレームセットは、

```json
{
    "writer": {
        "iss": "https://idp.example.org",
        "sub": "D5CF3F49E5E731C3"
    }
}
```


## 6. レスポンスの受け取り。

特別な処理は無し。


## 7. レスポンスの転送

リクエストがセッションを宣言せず、関連するアカウントが処理の主体のみだった場合、セッションを発行する。
セッションにアカウントタグ、アクセストークン、アカウント情報、要請元 TA の ID を紐付ける。
セッションを、Cookie の Edo-Cooperation で通知する。


## 8. エラーレスポンス

[TA 間連携プロトコル] を参照のこと。


## 9. 外部データ

以下に分ける。

* 共有データ
    * 他のプログラムと共有する可能性のあるもの。
* 非共有データ
    * 共有するとしてもこのプログラムの別プロセスのみのもの。


### 9.1. 共有データ


### 9.2. 非共有データ


#### 9.2.1. セッション

以下を含む。

* ID
* 有効期限
* 処理の主体のアカウントタグ
* 処理の主体のアクセストークン
    * ID
    * タグ
    * 有効期限
* 処理の主体の情報
* 要請元 TA の ID

以下の操作が必要。

* 保存
* ID による取得


<!-- 参照 -->
[JWT]: https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-32
[TA 間連携プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/ta_cooperation.md
