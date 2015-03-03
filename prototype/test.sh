#!/bin/bash

luajit=../../lib/lua-jit/bin/luajit

function test() {
    test_script=$1

    echo -e "\033[32m$test_script testing...\033[0m"
    if $luajit $test_script; then
        echo -e "\033[32m$test_script test pass!\033[0m"
    else
        echo -e "\033[31m$test_script test fail!\033[0m"
    fi
}

# test lib
cd lua/test
test lib/cookie_manager_test.lua
test lib/exiter_test.lua
test lib/redis_mock_test.lua
test lib/token_manager_test.lua

# test http
test http/login_test.lua
test http/callback_test.lua
test http/add_header_test.lua
test http/decrypt_test.lua
