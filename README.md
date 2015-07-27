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

ユーザーの認証、TA 間連携の受け入れを代行する。

セッションに基づく通過判断のみ行うフロントエンドの [nginx] モジュールと、セッション未確立時に代行処理を行うバックエンドのサーバープログラムからなる。


## 1. インストール


### 1.1. フロントエンド

script/install.sh で nginx ごとセットアップする。

```shell
./script/install.sh
```

インストール先を変える場合は install_dir で指定する。

```shell
install_dir=/home/edo-auth ./script/install.sh
```


#### 1.1.1. インストールオプション

|シェル変数|初期値|値|
|:--|:--|:--|
|install_dir|リポジトリ直下の root|インストール先のパス|


#### 1.1.2. 既存の nginx から使う

必要なモジュールと共に nginx がコンパイルされているなら、
リポジトリ直下の lua ファイルを nginx の prefix 直下の lua ディレクトリに、
lib 以下の lua ファイルを lua/lib ディレクトリにコピーする。


### 1.2. バックエンド

[go] が必要。
go のインストールは http://golang.org/doc/install を参照のこと。

go をインストールしたら、

```shell
go get github.com/realglobe-Inc/edo-auth
```

適宜、依存ライブラリを `go get` すること。


## 2. 実行


### 2.1. DB の準備

キャッシュやセッション用に [redis]、ID プロバイダ情報用に [mongodb] が必要になる。

mongodb への ID プロバイダ情報の同期は別口で行う。


### 2.2. フロントエンド


#### 2.2.1. 設定

[sample/nginx.conf](sample/nginx.conf) を参考に設定する。

オプションは nginx 変数で指定する。

|オプション変数名|初期値|値|
|:--|:--|:--|
|$edo\_backend\_location|モジュール依存|バックエンドを指す nginx の Location|
|$edo\_log\_level|debug|edo-auth のログを nginx のどのログレベルで出力するか|
|$edo\_redis\_address|127.0.0.1:6379|使用する redis のアドレス|


#### 2.2.2. 起動

nginx の起動方法に従う。

install.sh で /home/edo-auth にインストールしたとすれば、

```shell
/home/edo-auth/opt/nginx/sbin/nginx
```


### 2.3. バックエンド


#### 2.3.1. 起動

単独で実行できる。

```shell
${GOPATH}/bin/edo-auth
```


#### 2.3.2. 起動オプション

|オプション名|初期値|値|
|:--|:--|:--|
|-noVeri|`false`|通信先の SSL 証明書を検証しないかどうか|

その他は `-h` で確認すること。


#### 2.3.3. デーモン化

単独ではデーモンとして実行できないため、[Supervisor] 等と組み合わせて行う。


## 3. 動作仕様


### 3.1. 概要

通過した通信がユーザー認証済み、または、正しく TA 間連携が行われていることを保証する。


### 3.2. 詳細

以下を参照。

* [ユーザー認証手続き代行フロントエンド](/SPEC_auth.md)
* [ユーザー認証手続き代行バックエンド](/api/coop/README.md)
* [TA 間連携受け入れ代行フロントエンド](/SPEC_coop.md)
* [TA 間連携受け入れ代行バックエンド](/page/auth/README.md)


## 4. ライセンス

Apache License, Version 2.0


<!-- 参照 -->
[Supervisor]: http://supervisord.org/
[go]: http://golang.org/
[mongodb]: https://www.mongodb.org/
[nginx]: http://nginx.org/
[redis]: http://redis.io/
