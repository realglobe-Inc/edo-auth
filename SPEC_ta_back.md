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


# edo-auth TA バックエンドの仕様（目標）

以降の動作記述において、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


## 1. 概要

1. フロントエンドから仲介コードを受け取る。
2. 全ての仲介コードについて次の 1, 2 を行う。
    1. 仲介コードを発行した IdP に、仲介コードを送る。
    2. IdP から、仲介情報を受け取る。
3. フロントエンドに仲介情報を返す。

```
+--------+                    +--------+
|        |----(1) codes------>|        |                      +-------+
|        |                    |        |----(2-1) code------->|       |
|edo auth|                    |edo auth|                      |  IdP  |
|TA      |                    |TA      |----(2-2) userinfo--->|       |
|frontend|                    |backend |                      +-------+
|        |                    |        |                         ...
|        |                    |        |
|        |<---(5) userinfo----|        |
+--------+                    +--------+
```


## 2. リクエストの受け取り

[TA 間連携プロトコル]も参照のこと。


## 3. IdP への仲介リクエスト

[TA 間連携プロトコル]を参照のこと。


## 4. IdP からの仲介情報受け取り

[TA 間連携プロトコル]を参照のこと。

1 つでもエラーであれば、そのエラーを返す。


## 5. 仲介情報の返送

以下の HTTP ヘッダにて仲介情報を返す。

* X-Auth-User
    * 処理の主体の情報をクレームセットに含む `alg` が `none` な [JWT]。
* X-Auth-User-Tag
    * 処理の主体のアカウントタグ。
* X-Auth-Related-Users
    * 処理の主体でないアカウントが含まれる場合のみ。
      処理の主体でないアカウント全てについて、アカウントタグからアカウント情報へのマップをクレームセットに含む `alg` が `none` な [JWT]。
* X-Auth-From-Id
    * 連携元 TA の ID。

また、[TA 間連携プロトコル]のセッションが利用可能な場合はセッションを Set-Cookie ヘッダで通知する。


### 5.1. 仲介情報の返送例

```http
HTTP/1.1 200 OK
X-Auth-User: eyJhbGciOiJub25lIn0.eyJhdF9leHAiOjE0MjY1NjEyNjIsImF0X3RhZyI6InVudG5G
    ZHhOMDMiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsInN1YiI6Ijc1NUI2MjkyMDhF
    REZEQzIifQ.
X-Auth-User-Tag: reader
X-Auth-Related-Users: eyJhbGciOiJub25lIn0.eyJ3cml0ZXIiOnsiaXNzIjoiaHR0cHM6Ly9pZHA
    uZXhhbXBsZS5vcmciLCJzdWIiOiJENUNGM0Y0OUU1RTczMUMzIn19.
X-Auth-From-Id: https://from.example.org
```


## 6. エラーレスポンス

[TA 間連携プロトコル] を参照のこと。


## 7. 外部データ

以下に分ける。

* 共有データ
    * 他のプログラムと共有する可能性のあるもの。
* 非共有データ
    * 共有するとしてもこのプログラムの別プロセスのみのもの。


### 7.1. 共有データ


#### 7.1.1. IdP 情報

以下を含む。

* ID
* 検証鍵
* 連携先仲介エンドポイント

以下の操作が必要。

* ID による取得


#### 7.1.2. アクセストークン

以下を含む。

* ID
* タグ
* 有効期限
* 発行 IdP の ID
* 許可スコープ

以下の操作が必要。

* 保存


### 7.2. 非共有データ


<!-- 参照 -->
[JWT]: https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-32
[TA 間連携プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/ta_cooperation.md
