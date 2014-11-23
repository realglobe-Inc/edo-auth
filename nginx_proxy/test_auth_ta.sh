#!/bin/sh -e

redis_port=${redis_port:=6379}
nginx_port=${nginx_port:=7000}
nginx_ver=${nginx_ver:=1.7.4}

# デバッグ用。
redis_num () {
    netstat -an | grep " 127.0.0.1:${redis_port} " | wc -l
}

del_nginx () {
    for i in $@; do
        ${i}/sbin/nginx -p ${i} -s stop
        #rm -rf ${i}
    done
}

(cd $(dirname $0)
    if ! nc -z localhost $redis_port; then
        ./lib/redis/bin/redis-server - <<EOF
daemonize yes
port $redis_port
EOF
    fi

    # プロキシ先を立てる。
    while true; do
        if ! nc -z localhost $nginx_port; then
            rm -rf /tmp/edo-auth-dest
            cp -r lib/nginx_${nginx_ver} /tmp/edo-auth-dest
            cat <<EOF > /tmp/edo-auth-dest/conf/nginx.conf
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    types {
        text/plain txt;
        text/html  html;
    }
    default_type  text/plain;

    sendfile        on;

    keepalive_timeout  65;

    server {
        listen       ${nginx_port};
        server_name  localhost;

        location / {
            root   html;
            index  index.html index.htm;
        }

        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   html;
        }
    }
}
EOF
            /tmp/edo-auth-dest/sbin/nginx -p /tmp/edo-auth-dest
            trap 'del_nginx /tmp/edo-auth-dest' EXIT
            break
        else
            nginx_port=$((${nginx_port} + 1))
        fi
    done
    dest_nginx_port=${nginx_port}
    echo "start destination nginx at port ${dest_nginx_port}"

    while true; do
        if ! nc -z localhost $nginx_port; then
            rm -rf /tmp/edo-auth
            cp -r lib/nginx_${nginx_ver} /tmp/edo-auth
            cat <<EOF > /tmp/edo-auth/conf/nginx.conf
worker_processes  2;

events {
    worker_connections  1024;
}

http {
    types {
        text/html  html;
        text/plain txt;
    }
    default_type  text/plain;

    sendfile        on;

    keepalive_timeout  65;

    server {
        listen       ${nginx_port};
        server_name  localhost;

        location / {
            more_clear_headers "X-Edo-Ta-Id";
            more_clear_headers "X-Forwarded-For"; # 手前に別のリバースプロキシがいるならこの設定は消す。

            set \$edo_module_dir $(pwd);
            set \$edo_auth_log_level error; # デバッグ。
            set \$edo_auth_public_key_directory \$edo_module_dir/sample/public_key;
            access_by_lua_file \$edo_module_dir/lua/auth_ta.lua;

            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header Host \$http_host;

            proxy_pass http://localhost:${dest_nginx_port}/;
        }

        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   html;
        }
    }
}
EOF
            /tmp/edo-auth/sbin/nginx -p /tmp/edo-auth
            trap 'del_nginx /tmp/edo-auth-dest /tmp/edo-auth' EXIT
            break
        else
            nginx_port=$((${nginx_port} + 1))
        fi
    done

    echo "start edo-auth at port ${nginx_port}"


    # 未認証。
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '< Set-Cookie:.* \+X-Edo-Auth-Ta-Session=' ${TMP_FILE}; then
        echo "Error (unauthenticated): no X-Edo-Auth-Ta-Session in Cookie" 1>&2
        exit 1
    elif ! grep -q '^< X-Edo-Auth-Ta-Token:' ${TMP_FILE}; then
        echo "Error (unauthenticated): no X-Edo-Auth-Ta-Token" 1>&2
        exit 1
    elif ! grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (unauthenticated): no X-Edo-Auth-Ta-Error" 1>&2
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
    SIGN=$(printf ${TOKEN} | openssl dgst -${HASH} -binary | openssl pkeyutl -sign -inkey sample/private_key/${TA}.key -pkeyopt digest:${HASH} | base64 | tr -d '\n')

    ./lib/redis/bin/redis-cli -p ${redis_port} del "session:authenticated:${SESSION}" > /dev/null
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" \
        -H "X-Edo-Auth-Ta-Id: ${TA}" \
        -H "X-Edo-Auth-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Auth-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (authenticating): "$(grep '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}) 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 200 OK' ${TMP_FILE}; then
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
    curl -v --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (authenticated): "$(grep '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}) 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 200 OK' ${TMP_FILE}; then
        echo "Error (authenticated): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    ./lib/redis/bin/redis-cli -p ${redis_port} del "session:authenticated:${SESSION}" > /dev/null
    echo "----- OK: response to authenticated request -----"


    # セッション相手のアドレスが違っていたら 403 Forbidden を返すか。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:authenticated:${SESSION}" 10 '{"auth":true,"id":"'${SESSION}'","ta":"12345","client":"192.0.2.1"}' > /dev/null

    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (different source): no X-Edo-Auth-Ta-Error" 1>&2
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
        --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" \
        -H "X-Edo-Auth-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Auth-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (no X-Edo-Auth-Ta-Id): no X-Edo-Auth-Ta-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (no X-Edo-Auth-Ta-Id): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" \
        -H "X-Edo-Auth-Ta-Id: ${TA}" \
        -H "X-Edo-Auth-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (no X-Edo-Auth-Ta-Token-Sign): no X-Edo-Auth-Ta-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (no X-Edo-Auth-Ta-Token-Sign): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response in lack of information -----"


    # 公開鍵が無かったら 403 Forbidden を返すか。
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" \
        -H "X-Edo-Auth-Ta-Id: ${TA}-dewa-nai" \
        -H "X-Edo-Auth-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Auth-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (no public key): no X-Edo-Auth-Ta-Error" 1>&2
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
        --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" \
        -H "X-Edo-Auth-Ta-Id: ${TA}" \
        -H "X-Edo-Auth-Ta-Token-Sign: "$(printf ${TOKEN}F | openssl dgst -${HASH} -binary | openssl pkeyutl -sign -inkey sample/private_key/${TA}.key -pkeyopt digest:${HASH} | base64 | tr -d '\n') \
        -H "X-Edo-Auth-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if ! grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (invalid sign): no X-Edo-Auth-Ta-Error" 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 403 Forbidden' ${TMP_FILE}; then
        echo "Error (invalid sign): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response in invalid sign -----"


    # 公開鍵が証明書でも大丈夫か。
    TA=test
    ./lib/redis/bin/redis-cli -p ${redis_port} setex "session:unauthenticated:${SESSION}" 10 '{"id":"'${SESSION}'","token":"'${TOKEN}'","client":"127.0.0.1"}' > /dev/null
    SIGN=$(printf ${TOKEN} | openssl dgst -${HASH} -binary | openssl pkeyutl -sign -inkey sample/private_key/${TA}.key -pkeyopt digest:${HASH} | base64 | tr -d '\n')
    TMP_FILE=/tmp/$(basename ${0%.*})$(date +"%y%m%d%H%M%S%N")
    curl -v \
        --cookie "X-Edo-Auth-Ta-Session"="${SESSION}" \
        -H "X-Edo-Auth-Ta-Id: ${TA}" \
        -H "X-Edo-Auth-Ta-Token-Sign: ${SIGN}" \
        -H "X-Edo-Auth-Hash-Function: ${HASH}" \
        http://localhost:${nginx_port}/ > ${TMP_FILE} 2>&1
    if grep -q '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}; then
        echo "Error (authenticating): "$(grep '^< X-Edo-Auth-Ta-Error:' ${TMP_FILE}) 1>&2
        exit 1
    elif ! grep -q '^< HTTP/[0-9.]\+ 200 OK' ${TMP_FILE}; then
        echo "Error (authenticating): invalid status "$(grep '^< HTTP/[0-9.]\+ ' ${TMP_FILE}) 1>&2
        exit 1
    elif [ -n ""$(./lib/redis/bin/redis-cli -p ${redis_port} get "session:unauthenticated:${SESSION}") ]; then
        echo "Error (authenticating): unauthentcated session remains" $(./lib/redis/bin/redis-cli -p ${redis_port} get "session:unauthenticated:${SESSION}") 1>&2
        exit 1
    fi
    rm ${TMP_FILE}

    echo "----- OK: response in public key in certification -----"


    # redis との接続が再利用されているか。
    if [ $(netstat -an | grep " 127.0.0.1:${redis_port} .* ESTABLISHED" | wc -l) != 2 ]; then
        echo "Error: not one redis socket" 1>&2
        exit 1
    fi

    echo "----- OK: redis connection -----"
)
