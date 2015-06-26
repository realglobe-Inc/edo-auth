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


# nginx -s reload が反映されるのを sleep して待つ期間。
INTERVAL=0.5

PROJECT_DIR=$(cd $(dirname $0)/.. && pwd)

WORK_DIR=${WORK_DIR:=/tmp/edo-auth-test}
NGINX_DIR=${NGINX_DIR:=${PROJECT_DIR}/root/opt/nginx}
REDIS_SERVER=${REDIS_SERVER:=$(which redis-server)}
REDIS_CLIENT=${REDIS_CLIENT:=$(which redis-cli)}

if ! [ -f ${NGINX_DIR}/sbin/nginx ]; then
    echo "no nginx in ${NGINX_DIR}" 1>&2
    exit 1
elif ! [ -f ${REDIS_SERVER} ]; then
    echo "no redis at ${REDIS_SERVER}" 1>&2
    exit 1
fi

NGINX_PORT=${NGINX_PORT:=7000}
while nc -z localhost ${NGINX_PORT}; do
    NGINX_PORT=$((${NGINX_PORT} + 1))
done
REDIS_PORT=${REDIS_PORT:=$((${NGINX_PORT} + 1))}
while nc -z localhost ${REDIS_PORT}; do
    REDIS_PORT=$((${REDIS_PORT} + 1))
done


if ! [ -d ${WORK_DIR} ]; then
    mkdir -p ${WORK_DIR}
    echo "${WORK_DIR} was created"
fi

(cd ${WORK_DIR}

 nginx_prefix=${WORK_DIR}/nginx
 mkdir -p ${nginx_prefix}/conf
 mkdir -p ${nginx_prefix}/logs
 mkdir -p ${nginx_prefix}/lua/lib
 mkdir -p ${nginx_prefix}/lua/test
 cp ${PROJECT_DIR}/*.lua ${nginx_prefix}/lua/
 cp ${PROJECT_DIR}/lib/*.lua ${nginx_prefix}/lua/lib/
 cp ${PROJECT_DIR}/test/*.lua ${nginx_prefix}/lua/test/


 cat <<EOF > redis.conf
daemonize yes
port $REDIS_PORT
EOF

 ${REDIS_SERVER} redis.conf
 close_script="${REDIS_CLIENT} -p ${REDIS_PORT} shutdown"
 trap "${close_script}" EXIT

 while ! nc -z localhost ${REDIS_PORT}; do
     sleep ${INTERVAL}
 done

 # redis が立った。

 cp -r ${PROJECT_DIR}/test ${nginx_prefix}/lua/
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            set \$redis_address 127.0.0.1:${REDIS_PORT};
            access_by_lua_file lua/test/redis.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix}
 close_script="${close_script}; ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s stop"
 trap "${close_script}" EXIT

 while ! nc -z localhost ${NGINX_PORT}; do
     sleep ${INTERVAL}
 done

 # nginx が立った。


 # ############################################################
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            access_by_lua_file lua/test/table.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== table passed ====="


 # ############################################################
 ${REDIS_CLIENT} -p ${REDIS_PORT} flushall > /dev/null
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            set \$redis_address 127.0.0.1:${REDIS_PORT};
            access_by_lua_file lua/test/redis_wrapper.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== redis passed ====="


 # ############################################################
 redis_path=$(pwd)/redis.sock
 cat <<EOF > redis.sock.conf
daemonize yes
unixsocket ${redis_path}
EOF
 ${REDIS_SERVER} redis.sock.conf
 close_script="${close_script}; ${REDIS_CLIENT} -s ${redis_path} shutdown"
 trap "${close_script}" EXIT
 while ! [ -e ${redis_path} ]; do
     sleep ${INTERVAL}
 done

 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            set \$redis_address unix:${redis_path};
            access_by_lua_file lua/test/redis_wrapper.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== redis (unix socket) passed ====="


 # ############################################################
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            access_by_lua_file lua/test/auth_session.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== user session passed ====="


 # ############################################################
 ${REDIS_CLIENT} -p ${REDIS_PORT} flushall > /dev/null
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            set \$redis_address 127.0.0.1:${REDIS_PORT};
            access_by_lua_file lua/test/auth_session_db.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== user session DB passed ====="


 # ############################################################
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            access_by_lua_file lua/test/coop_session.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== TA session passed ====="


 # ############################################################
 ${REDIS_CLIENT} -p ${REDIS_PORT} flushall > /dev/null
 cat <<EOF > ${nginx_prefix}/conf/nginx.conf
events {}
http {
    lua_package_path '\${prefix}lua/?.lua;;';
    server {
        listen       ${NGINX_PORT};
        location / {
            set \$redis_address 127.0.0.1:${REDIS_PORT};
            access_by_lua_file lua/test/coop_session_db.lua;
        }
    }
}
EOF
 ${NGINX_DIR}/sbin/nginx -p ${nginx_prefix} -s reload
 sleep ${INTERVAL}

 result=$(curl -o out -s -w "%{http_code}" http://localhost:${NGINX_PORT})
 if [ "${result}" != "200" ]; then
     echo ${result} 1>&2
     cat out 1>&2
     exit 1
 fi
 echo "===== TA session DB passed ====="
)

echo "===== all test passed ====="
