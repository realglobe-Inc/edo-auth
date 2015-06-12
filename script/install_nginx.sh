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


# Lua に対応した nginx をインストールする。


headers_more_ver=${headers_more_ver:=v0.26}
lua_nginx_ver=${lua_nginx_ver:=v0.9.15}
luajit_ver=${luajit_ver:=2.0.3}
nginx_ver=${nginx_ver:=1.7.10}
ngx_devel_kit_ver=${ngx_devel_kit_ver:=v0.2.19}
openssl_ver=${openssl_ver:=1.0.2b}
pcre_ver=${pcre_ver:=8.36}
zlib_ver=${zlib_ver:=1.2.8}

project_dir=$(cd $(dirname $0)/.. && pwd)

install_dir=${install_dir:=${project_dir}/root}
# フルパスにする。
if ! [ -d ${install_dir} ]; then
    mkdir -p ${install_dir}
    echo "${install_dir} was created"
fi
install_dir=$(cd ${install_dir} && pwd)
src_dir=${install_dir}/src
nginx_dir=${nginx_dir:=${install_dir}/opt/nginx}


mkdir -p ${src_dir}

(cd ${src_dir}

 # zlib
 if ! [ -d zlib-${zlib_ver} ]; then
     if ! [ -f zlib-${zlib_ver}.tar.gz ]; then
         wget http://zlib.net/zlib-${zlib_ver}.tar.gz
     fi
     tar zxf zlib-${zlib_ver}.tar.gz
 fi

 # pcre
 if ! [ -d pcre-${pcre_ver} ]; then
     if ! [ -f pcre-${pcre_ver}.tar.gz ]; then
         wget http://downloads.sourceforge.net/project/pcre/pcre/${pcre_ver}/pcre-${pcre_ver}.tar.gz
     fi
     tar zxf pcre-${pcre_ver}.tar.gz
 fi

 # openssl
 if ! [ -d openssl-${openssl_ver} ]; then
     if ! [ -f openssl-${openssl_ver}.tar.gz ]; then
         wget ftp://ftp.openssl.org/source/openssl-${openssl_ver}.tar.gz
     fi
     tar zxf openssl-${openssl_ver}.tar.gz
 fi

 # headers-more
 if ! [ -d headers-more-nginx-module ]; then
     git clone https://github.com/agentzh/headers-more-nginx-module.git
 fi
 (cd headers-more-nginx-module/
  git fetch
  git checkout ${headers_more_ver}
 )

 # LuaJIT
 if ! [ -f ${install_dir}/bin/luajit ]; then
     if ! [ -f LuaJIT-${luajit_ver}.tar.gz ]; then
         wget http://luajit.org/download/LuaJIT-${luajit_ver}.tar.gz
     fi
     if ! [ -d LuaJIT-${luajit_ver} ]; then
         tar zxf LuaJIT-${luajit_ver}.tar.gz
     fi
     (cd LuaJIT-${luajit_ver}
      make clean
      make PREFIX=${install_dir}
      make install PREFIX=${install_dir}
     )
 fi

 # ngx_devel_kit
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
 if ! [ -f ${nginx_dir}/sbin/nginx ]; then
     if ! [ -d nginx-${nginx_ver} ]; then
         if ! [ -f nginx-${nginx_ver}.tar.gz ]; then
             wget http://nginx.org/download/nginx-${nginx_ver}.tar.gz
         fi
         tar zxf nginx-${nginx_ver}.tar.gz
     fi
     (cd nginx-${nginx_ver}
      export LUAJIT_LIB=${install_dir}/lib
      export LUAJIT_INC=${install_dir}/include/luajit-2.0
      if [ -f Makefile ]; then
          make clean
      fi
      ./configure \
          --prefix=${nginx_dir} \
          --with-http_gzip_static_module \
          --with-http_realip_module \
          --with-http_spdy_module \
          --with-http_ssl_module \
          --with-http_sub_module \
          --with-ld-opt="-Wl,-rpath=${install_dir}/lib" \
          --with-openssl=${src_dir}/openssl-${openssl_ver} \
          --with-pcre=${src_dir}/pcre-${pcre_ver} \
          --with-zlib=${src_dir}/zlib-${zlib_ver} \
          --add-module=${src_dir}/headers-more-nginx-module \
          --add-module=${src_dir}/ngx_devel_kit \
          --add-module=${src_dir}/lua-nginx-module
      make
      make install
     )
 fi
)


echo "---------- nginx was installed in ${install_dir} ----------"
