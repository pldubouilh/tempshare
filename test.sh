#!/bin/bash
set -xe

function error_exit() {
    echo_red "\n==> TESTS FAILED!"
}
trap error_exit ERR

function echo_green() {
    printf "\e[32m$1\e[0m\n"
}

function echo_red() {
    printf "\e[31m$1\e[0m\n"
}

function cleanup() {
    set +ex
    rm -rf $temp
    kill $(jobs -p)
}
trap cleanup EXIT

temp=$(mktemp -d)

# start server, get url
./tempshare test-fixture/smol > $temp/logs &
sleep 1
url=$(cat $temp/logs | grep -o 'http://[^ ]*')

# download and compare hashes
curl -s $url > $temp/output
diff $temp/output test-fixture/smol

# one more time
curl -s $url > $temp/output
diff $temp/output test-fixture/smol

# this should fail - server is stopped
curl -f $url 2>&1 | grep "Failed to connect"

# now test zips
./tempshare test-fixture > $temp/logs &
sleep 1
url=$(cat $temp/logs | grep -o 'http://[^ ]*')

# download and compare hashes
curl -s $url > $temp/output
cd $temp && unzip -q output
cd -
diff $temp/assets/verysmol test-fixture/assets/verysmol

echo_green "All tests passed"