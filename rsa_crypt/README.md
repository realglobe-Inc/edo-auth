## 事前準備

# RSA秘密鍵のパス
private_key_path=/path/to/private_key

# 秘密鍵の作成
openssl genrsa 2048 > ${private_key_path}

# 公開鍵の作成
openssl rsa -pubout < ${private_key_path} > /path/to/public_key

# BKFかどこかに公開鍵を登録し、他のサービスが鍵を取得できるようにしておく
# 多分APIかWeb UIか何かで公開鍵を登録すると、UUIDが発行される


## サービスにリクエストを投げる場合

# 暗号化に使用する秘密鍵の（公開鍵の登録時に発行された）UUID
private_key_uuid=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
# リクエストを送る対象のUUID
receiver_uuid=yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
# タイムスタンプのフォーマットは適当（後から変更するかも）
timestamp=`date +"%Y%m%d%H%M%S" --utc`
# 上の3つをカンマで区切り、秘密鍵で暗号化+base64エンコード+改行を削除する
encoded_string=`printf ${private_key_uuid},${receiver_uuid},${timestamp} | openssl rsautl -sign -inkey ${private_key_path} | base64 | tr -d '\n'`

# HTTPリクエスト
# 鍵のUUIDと暗号化した文字列をヘッダに追加すれば、受け取った側が上手い事やってくれるはず
curl https://edo-service.com/ -H "X-EDO-Private-Key-UUID: ${private_key_uuid}" -H "X-EDO-Auth-Encoded-Token: ${encoded_string}"
