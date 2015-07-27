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


# Lua に対応した nginx がインストール済みとして、
# Lua ライブラリをインストールする。


lua_cjson_ver=${lua_cjson_ver:=2.1.0}
lua_resty_redis_ver=${lua_resty_redis_ver:=688f932}
luajit_ver=${luajit_ver:=2.0.3}

project_dir=$(cd $(dirname $0)/.. && pwd)

install_dir=${install_dir:=${project_dir}/root}
# フルパスにする。
if ! [ -d ${install_dir} ]; then
    mkdir -p ${install_dir}
    echo "${install_dir} was created"
fi
install_dir=$(cd ${install_dir} && pwd)
src_dir=${install_dir}/src


mkdir -p ${src_dir}

(cd ${src_dir}

 # lua-resty-redis
 if ! [ -d ${install_dir}/share/lua/5.1/resty ]; then
     if ! [ -d lua-resty-redis ]; then
         git clone https://github.com/openresty/lua-resty-redis.git
     fi
     (cd lua-resty-redis
      git fetch
      git checkout ${lua_resty_redis_ver}
      cp -rf lib/resty ${install_dir}/share/lua/5.1/
     )
 fi

 # lua-cjson
 if ! [ -f ${install_dir}/lib/lua/5.1/cjson.so ];then
     if ! [ -d lua-cjson ]; then
         git clone https://github.com/mpx/lua-cjson.git
     fi
     (cd lua-cjson
      git fetch
      git checkout ${lua_cjson_ver}
      make clean
      make CFLAGS=-I${src_dir}/LuaJIT-${luajit_ver}/src
      make PREFIX=${install_dir} install
     )
 fi
)


echo "---------- Lua library was installed in ${install_dir} ----------"
