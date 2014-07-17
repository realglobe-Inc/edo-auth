#!/bin/bash

sso_module_dir=$HOME/edo-sso/oauth2_module
src_dir=$sso_module_dir/src
lib_dir=$sso_module_dir/lib

mkdir $src_dir
mkdir $lib_dir

redis_dir=$lib_dir/redis
luajit_dir=$lib_dir/lua-jit
nginx_dir=$lib_dir/nginx_1.7.2

# redis
cd $src_dir
git clone https://github.com/antirez/redis.git
cd redis
git checkout 2.8.12
make
make install PREFIX=$redis_dir
cd $redis_dir/bin
cp $sso_module_dir/redis.conf .
./redis-server redis.conf

# LuaJIT
cd $src_dir
wget http://luajit.org/download/LuaJIT-2.0.3.tar.gz
tar zxf LuaJIT-2.0.3.tar.gz
cd LuaJIT-2.0.3
make PREFIX=$luajit_dir
make install PREFIX=$luajit_dir

# lua-cjson
cd $src_dir
git clone https://github.com/efelix/lua-cjson.git
cd lua-cjson/
make CFLAGS=-I$src_dir/LuaJIT-2.0.3/src
make PREFIX=PREFIX=$luajit_dir install

## nginx modules
# pcre
cd $src_dir
wget ftp://ftp.csx.cam.ac.uk/pub/software/programming/pcre/pcre-8.33.tar.gz
tar zxf pcre-8.33.tar.gz

# openssl
cd $src_dir
wget http://www.openssl.org/source/openssl-1.0.1h.tar.gz
tar zxf openssl-1.0.1h.tar.gz

# headers-more
cd $src_dir
git clone https://github.com/agentzh/headers-more-nginx-module.git
cd headers-more-nginx-module/
git checkout v0.25

# redis2-nginx-module
cd $src_dir
git clone https://github.com/openresty/redis2-nginx-module.git
cd redis2-nginx-module/
git checkout v0.11

# ngx_devel_kit
cd $src_dir
git clone https://github.com/simpl/ngx_devel_kit.git
cd ngx_devel_kit/
git checkout v0.2.19

# lua-nginx-module
cd $src_dir
git clone https://github.com/openresty/lua-nginx-module.git
cd lua-nginx-module
git checkout v0.9.9

## nginx
cd $src_dir
wget http://nginx.org/download/nginx-1.7.2.tar.gz
tar zxf nginx-1.7.2.tar.gz
cd nginx-1.7.2

export LUAJIT_LIB=$luajit_dir/lib
export LUAJIT_INC=$luajit_dir/include/luajit-2.0
./configure\
 --prefix=$nginx_dir\
 --with-openssl=$src_dir/openssl-1.0.1h\
 --with-http_ssl_module\
 --with-http_gzip_static_module\
 --with-http_realip_module\
 --with-http_spdy_module\
 --with-pcre=$src_dir/pcre-8.33\
 --with-ld-opt="-Wl,-rpath=$luajit_dir/lib"\
 --add-module=$src_dir/headers-more-nginx-module\
 --add-module=$src_dir/redis2-nginx-module\
 --add-module=$src_dir/ngx_devel_kit\
 --add-module=$src_dir/lua-nginx-module
make
make install

mkdir $nginx_dir/conf/oauth2
cp $sso_module_dir/nginx.conf/nginx.conf.sample $nginx_dir/conf/oauth2/nginx.conf
