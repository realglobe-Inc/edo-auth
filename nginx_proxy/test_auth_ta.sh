#!/bin/sh -e

redis_port=6379
nginx_port=7000
nginx_ver=1.7.4

# デバッグ用。
redis_num () {
    netstat -an | grep " 127.0.0.1:${redis_port} " | wc -l
}


(cd $(dirname $0)
    if ! nc -z localhost $redis_port; then
        ./lib/redis/bin/redis-server - <<EOF
daemonize yes
port $redis_port
EOF
    fi

    if ! nc -z localhost $nginx_port; then
        ./lib/nginx_${nginx_ver}/sbin/nginx
    fi


    # 未認証。
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '< Set-Cookie:.* \+X-Edo-Ta-Session=' ${TMP_FILE}; then
        echo "Error (unauthenticated): no X-Edo-Ta-Session in Cookie" 1>&2
        exit 1
    elif ! grep -q '^< X-Edo-Ta-Token:' ${TMP_FILE}; then
        echo "Error (unauthenticated): no X-Edo-Ta-Token" 1>&2
        exit 1
    elif ! grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (unauthenticated): no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 401 Unauthorized' ${TMP_FILE}; then
        echo "Error (unauthenticated): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE})   1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response to unauthenticated request -----"


    # 認証中。
    TA="auth-ta-checker-no-id"
    SESSION="abcde"
    TOKEN="ABCDE"
    HASH="sha256"
    SIGN=$(printf ${TOKEN} | openssl dgst -${HASH} -binary | openssl pkeyutl -sign -inkey sample/private_key/${TA}.pem -pkeyopt digest:${HASH} | base64 | tr -d '\n')

    ./lib/redis/bin/redis-cli -p ${redis_port} del "session:authenticated:${SESSION}" > /dev/null
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Ta-Session"="${SESSION}" \
        -H "X-Edo-Ta-Id: ${TA}" \
        -H "X-Edo-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (authenticating): "$(grep '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}) 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ \(502 Bad Gateway\|200 OK\)' ${TMP_FILE}; then
        # 502 Bad Gateway はプロキシ先を用意してないとき。
        # プロキシ先が / に 200 OK を返す前提。
        echo "Error (authenticating): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    elif [ -n ""$(./lib/redis/bin/redis-cli -p ${redis_port} get "session:unauthenticated:${SESSION}") ]; then
        echo "Error (authenticating): unauthentcated session remains" $(./lib/redis/bin/redis-cli -p ${redis_port} get "session:unauthenticated:${SESSION}") 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    ./lib/redis/bin/redis-cli -p ${redis_port} del "session:authenticated:${SESSION}" > /dev/null
    echo "----- OK: response to authenticating request -----"


    # 認証済み。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:authenticated:${SESSION}" 10 '{"auth":true,"id":"'${SESSION}'","ta":"12345","client":"127.0.0.1"}' > /dev/null

    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v --cookie "X-Edo-Ta-Session"="${SESSION}" http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (authenticated): "$(grep '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}) 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ \(502 Bad Gateway\|200 OK\)' ${TMP_FILE}; then
        # 502 Bad Gateway はプロキシ先を用意してないとき。
        # プロキシ先が / に 200 OK を返す前提。
        echo "Error (authenticated): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    ./lib/redis/bin/redis-cli -p ${redis_port} del "session:authenticated:${SESSION}" > /dev/null
    echo "----- OK: response to authenticated request -----"


    # セッション相手のアドレスが違っていたら 403 Forbidden を返すか。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:authenticated:${SESSION}" 10 '{"auth":true,"id":"'${SESSION}'","ta":"12345","client":"192.0.2.1"}' > /dev/null

    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v --cookie "X-Edo-Ta-Session"="${SESSION}" http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (different source): no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (different source): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    ./lib/redis/bin/redis-cli -p ${redis_port} del "session:authenticated:${SESSION}" > /dev/null
    echo "----- OK: response to different source -----"


    # 認証情報が揃っていなかったら 403 Forbidden を返すか。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Ta-Session"="${SESSION}" \
        -H "X-Edo-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (no X-Edo-Ta-Id): no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (no X-Edo-Ta-Id): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Ta-Session"="${SESSION}" \
        -H "X-Edo-Ta-Id: ${TA}" \
        -H "X-Edo-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (no X-Edo-Ta-Token-Sign): no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (no X-Edo-Ta-Token-Sign): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response in lack of information -----"


    # 公開鍵が無かったら 403 Forbidden を返すか。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Ta-Session"="${SESSION}" \
        -H "X-Edo-Ta-Id: ${TA}-dewa-nai" \
        -H "X-Edo-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (no public key): no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (no public key): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response in no public key -----"


    # 署名がおかしかったら 403 Forbidden を返すか。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Ta-Session"="${SESSION}" \
        -H "X-Edo-Ta-Id: ${TA}" \
        -H "X-Edo-Ta-Token-Sign: "$(printf ${TOKEN}F | openssl dgst -${HASH} -binary | openssl pkeyutl -sign -inkey sample/private_key/${TA}.pem -pkeyopt digest:${HASH} | base64 | tr -d '\n') \
        -H "X-Edo-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Ta-Auth-Error:' ${TMP_FILE}; then
        echo "Error (invalid sign): no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (invalid sign): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response in invalid sign -----"


    # redis との接続が再利用されているか。
    if [ $(netstat -an | grep " 127.0.0.1:${redis_port} .* ESTABLISHED" | wc -l) != 2 ]; then
        echo "Error: not one redis socket" 1>&2
        exit 1
    fi

    echo "----- OK: redis connection -----"
)
