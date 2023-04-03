#!/usr/bin/env bash

set -e

nproc=${1:-"1"}

apt-get update

apt-get install -y make
apt-get install -y git 
apt-get install -y gcc g++
apt-get install -y fio


if go version &>/dev/null ; then
    apt install -y wget
    wget -c https://go.dev/dl/go1.20.2.linux-amd64.tar.gz
    tar -xzf go1.20.2.linux-amd64.tar.gz  -C /usr/local
    echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
    source ~/.profile
    mkdir -p ~/.config/go
    echo "
GOPROXY=https://goproxy.io,direct
GOPRIVATE=""
GOSUMDB=off
    " >> ~/.config/go/env
    source ~/.config/go/env
fi


# unminimize
# apt-get install man

# cd /tmp
# git clone https://github.com/axboe/liburing.git
# cd liburing
# ./configure --cc=gcc --cxx=g++
# make -j$(nproc)
# make install
# cd -
