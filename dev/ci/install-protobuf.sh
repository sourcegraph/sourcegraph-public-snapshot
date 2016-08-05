#!/bin/bash

set -ex

if [ -d ~/protobuf-3.0.2 ]; then
    exit 0
fi

cd ~
wget https://github.com/google/protobuf/archive/v3.0.2.tar.gz
tar -xzf v3.0.2.tar.gz
cd protobuf-3.0.2
./autogen.sh
./configure
make
cp src/.libs/lt-protoc src/protoc
