#!/bin/bash

set -ex

if [ -d ~/protobuf-3.1.0 ]; then
    exit 0
fi

cd ~
wget https://github.com/google/protobuf/archive/v3.1.0.tar.gz
tar -xzf v3.1.0.tar.gz
cd protobuf-3.1.0
./autogen.sh
./configure
make
cp src/.libs/lt-protoc src/protoc
