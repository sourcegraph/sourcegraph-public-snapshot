#!/bin/bash

set -ex

if [ -d ~/protobuf-3.0.0-beta-1 ]; then
    exit 0
fi

cd ~
wget https://github.com/google/protobuf/archive/v3.0.0-beta-1.tar.gz
tar -xzf v3.0.0-beta-1.tar.gz
cd protobuf-3.0.0-beta-1
# autogen script refers to googlecode which no longer exists, so we have to do terrible hacks.
curl -L -O https://github.com/google/googletest/archive/release-1.8.0.tar.gz
tar xf release-1.8.0.tar.gz && rm release-1.8.0.tar.gz
mv googletest-release-1.8.0/googletest/ .
mv googletest-release-1.8.0/googlemock/ gmock
rm -rf googletest-release-1.8.0
./autogen.sh
./configure
make
cp src/.libs/lt-protoc src/protoc
