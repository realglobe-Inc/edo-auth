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


script_dir=$(cd $(dirname $0) && pwd)
install_dir=${install_dir:=${script_dir}}
nginx=${nginx:=true}
nginx_port=${nginx_port:=7000}
proxy_pass=${proxy_pass:=http://example.org/}

# フルパスにする。
install_dir=$(cd ${install_dir} && pwd)

redis_ver=${redis_ver:=2.8.17}
luajit_ver=${luajit_ver:=2.0.3}
lua_redis_ver=${lua_redis_ver:=688f932}
lua_cjson_ver=${lua_cjson_ver:=2.1.0}
openssl_ver=${openssl_ver:=1.0.1j}
lua_openssl_ver=${lua_openssl_ver:=3494b8a}
headers_more_ver=${headers_more_ver:=v0.25}
ngx_devel_kit_ver=${ngx_devel_kit_ver:=v0.2.19}
lua_nginx_ver=${lua_nginx_ver:=v0.9.13}
nginx_ver=${nginx_ver:=1.7.7}

lua_dir=$(dirname $0)/lua
src_dir=${install_dir}/src
lib_dir=${install_dir}/lib
nginx_dir=${install_dir}/nginx

luajit_dir=${lib_dir}/lua-jit

mkdir -p ${src_dir}
mkdir -p ${lib_dir}
if ! [ -d ${install_dir}/lua ]; then
    cp -r ${lua_dir} ${install_dir}/
fi

(cd ${src_dir}

 # LuaJIT
 if ! [ -f LuaJIT-${luajit_ver}.tar.gz ]; then
     wget http://luajit.org/download/LuaJIT-${luajit_ver}.tar.gz
 fi
 if ! [ -d LuaJIT-${luajit_ver} ]; then
     tar zxf LuaJIT-${luajit_ver}.tar.gz
 fi
 if ! [ -d ${luajit_dir} ]; then
     (cd LuaJIT-${luajit_ver}
      make clean
      make PREFIX=${luajit_dir}
      make install PREFIX=${luajit_dir}
     )
 fi

 # lua-resty-redis
 if ! [ -d lua-resty-redis ]; then
     git clone https://github.com/openresty/lua-resty-redis.git
 fi
 (cd lua-resty-redis
  git fetch
  git checkout ${lua_redis_ver}
  if ! [ -d ${luajit_dir}/share/lua/5.1/resty ]; then
      cp -rf lib/resty ${luajit_dir}/share/lua/5.1/
  fi
 )

 # lua-cjson
 if ! [ -d lua-cjson ]; then
     git clone https://github.com/mpx/lua-cjson.git
 fi
 (cd lua-cjson
  git fetch
  git checkout ${lua_cjson_ver}
  if [ -z $(find ${luajit_dir} -path "*/cjson.so") ];then
      make clean
      make CFLAGS=-I${src_dir}/LuaJIT-${luajit_ver}/src
      make PREFIX=${luajit_dir} install
  fi
 )

 # openssl
 if ! [ -f openssl-${openssl_ver}.tar.gz ]; then
     wget http://www.openssl.org/source/openssl-${openssl_ver}.tar.gz
 fi
 if ! [ -d openssl-${openssl_ver} ]; then
     tar zxf openssl-${openssl_ver}.tar.gz
 fi
 if ! [ -d ${lib_dir}/openssl/ssl ]; then
     (cd openssl-${openssl_ver}
      ./config --prefix=${lib_dir}/openssl/ssl -fPIC shared zlib-dynamic threads
      make clean
      make
      make install
     )
 fi

 # lua-openssl
 if ! [ -d lua-openssl ]; then
     git clone https://github.com/zhaozg/lua-openssl.git
 fi
 (cd lua-openssl/
  git fetch
  git checkout ${lua_openssl_ver}
  git submodule init
  git submodule update
  if [ -z $(find ${luajit_dir} -path "*/openssl.so") ]; then
      make clean
      set +e # 最後の chcon が失敗するだけ。
      make PREFIX=${lib_dir}/openssl CFLAGS="-I${src_dir}/LuaJIT-${luajit_ver}/src -DPTHREADS -fPIC"
      set -e
      LUAV=5.1 make install PREFIX=${luajit_dir}
  fi
 )

 # pcre
 if ! [ -f pcre-8.33.tar.gz ]; then
     wget http://downloads.sourceforge.net/project/pcre/pcre/8.33/pcre-8.33.tar.gz
 fi
 if ! [ -d pcre-8.33 ]; then
     tar zxf pcre-8.33.tar.gz
 fi

 # headers-more
 if ! [ -d headers-more-nginx-module ]; then
     git clone https://github.com/agentzh/headers-more-nginx-module.git
 fi
 (cd headers-more-nginx-module/
  git fetch
  git checkout ${headers_more_ver}
 )

 # ngx_devel_kit
 # lua-nginx-module に必要。
 if ! [ -d ngx_devel_kit ]; then
     git clone https://github.com/simpl/ngx_devel_kit.git
 fi
 (cd ngx_devel_kit/
  git fetch
  git checkout ${ngx_deve_kit_ver}
 )

 # lua-nginx-module
 if ! [ -d lua-nginx-module ]; then
     git clone https://github.com/openresty/lua-nginx-module.git
 fi
 (cd lua-nginx-module
  git fetch
  git checkout ${lua_nginx_ver}
 )

 # nginx
 if ! [ -f nginx-${nginx_ver}.tar.gz ]; then
     wget http://nginx.org/download/nginx-${nginx_ver}.tar.gz
 fi
 if ! [ -d nginx-${nginx_ver} ]; then
     tar zxf nginx-${nginx_ver}.tar.gz
 fi
 if ! [ -d ${nginx_dir} ]; then
     (cd nginx-${nginx_ver}
      if [ -f Makefile ]; then
          make clean
      fi
      export LUAJIT_LIB=${luajit_dir}/lib
      export LUAJIT_INC=${luajit_dir}/include/luajit-2.0 # LuaJIT が 2.0.x じゃなくなったら変更？
      ./configure \
          --prefix=${nginx_dir} \
          --with-http_gzip_static_module \
          --with-http_realip_module \
          --with-http_spdy_module \
          --with-http_ssl_module \
          --with-http_sub_module \
          --with-ld-opt="-Wl,-rpath=${luajit_dir}/lib" \
          --with-openssl=${src_dir}/openssl-${openssl_ver} \
          --with-pcre=$src_dir/pcre-8.33 \
          --add-module=${src_dir}/headers-more-nginx-module \
          --add-module=${src_dir}/ngx_devel_kit \
          --add-module=${src_dir}/lua-nginx-module
      make
      make install

      sed 's/^\([ ]\+listen[ ]\+\)7000;$/\1'${nginx_port}';/' ${script_dir}/sample/nginx.conf | \
          sed 's/^\([ ]\+proxy_pass[ ]\+\)http:\/\/example\.org\/;$/\1'$(echo ${proxy_pass} | sed 's/\./\\\./g' | sed 's/\//\\\//g')';/' > \
              ${nginx_dir}/conf/nginx.conf
     )
 fi
)


if ${nginx} && ! nc -z localhost ${nginx_port}; then
    ${nginx_dir}/sbin/nginx
fi

echo "---------- Finished ----------"
