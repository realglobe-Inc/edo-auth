#!/bin/sh -e

# Copyright 2015 realglobe, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# sample/public_keys/auth-ta-checker.pub を登録する必要あり。

nginx_port=${nginx_port:=7000}
nginx_url=${nginx_url:=http://localhost:${nginx_port}/}

echo "----- Response of unauthorized request -----"
curl ${nginx_url}

curl -v ${nginx_url} > /tmp/edo-auth-test-file 2>&1

token=$(grep "X-Edo-Auth-Ta-Token" /tmp/edo-auth-test-file | tr -d '\r' | cut -d " " -f 3)
session=$(grep "X-Edo-Auth-Ta-Session" /tmp/edo-auth-test-file | cut -d " " -f 3 | cut -d "=" -f 2 | sed 's/;$//')
sign=$(printf ${token} | \
              openssl dgst -sha256 -binary | \
              openssl pkeyutl -sign -inkey $(dirname $0)/sample/private_keys/auth-ta-checker.key -pkeyopt digest:sha256 | \
              base64 | \
              tr -d '\n')

echo
echo
echo "----- Response of authorized request -----"
curl ${nginx_url}/ \
     -H "Cookie: X-Edo-Auth-Ta-Session=${session}" \
     -H "X-Edo-Auth-Ta-Id: auth-ta-checker" \
     -H "X-Edo-Auth-Ta-Token-Sign: ${sign}" \
     -H "X-Edo-Auth-Hash-Function: sha256"
