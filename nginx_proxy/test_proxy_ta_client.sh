#!/bin/sh -e

redis_port=6379
nginx_port=7000

(cd $(dirname $0)
    if ! nc -z localhost $redis_port; then
        ./lib/redis/bin/redis-server - <<EOF
daemonize yes
port $redis_port
EOF
    fi
    if ! nc -z localhost $nginx_port; then
        ./lib/nginx-${nginx_ver}/sbin/nginx
    fi

    # 未認証。
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! cat ${TMP_FILE} | grep -q '< Set-Cookie: \+X-Edo-Ta-Session *='; then
        echo "Error: no X-Edo-Ta-Session in Cookie" 1>&2
        exit 1
    elif ! cat ${TMP_FILE} | grep -q 'X-Edo-Ta-Token: '; then
        echo "Error: no X-Edo-Ta-Token" 1>&2
        exit 1
    elif ! cat ${TMP_FILE} | grep -q 'X-Edo-Ta-Auth-Error: '; then
        echo "Error: no X-Edo-Ta-Auth-Error" 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    # TODO 認証。

    # 認証済み。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:authenticated:abcde" 10 \
        '{"auth":true,"id":"abcde","ta":"12345","client":"127.0.0.1"}' > /dev/null

    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v -H "Cookie: X-Edo-Ta-Session=abcde" http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! cat ${TMP_FILE} | grep -q -e '< HTTP/[0-9.]\+ 502 Bad Gateway' -e "HTTP/[0-9.]\+ 200 OK"; then
        echo "Error: invalid status" 1>&2
        exit 1
    fi
    rm ${TMP_FILE}
)

echo "----- OK -----"
