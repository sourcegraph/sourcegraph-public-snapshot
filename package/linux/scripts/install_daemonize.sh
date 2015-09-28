#!/bin/bash

set -e

URL=https://github.com/bmc/daemonize/archive/release-1.7.6.tar.gz
RELEASE_DIR=daemonize-release-1.7.6
TARGET_DIR="$1"

pushd /tmp
curl -L ${URL} | tar zx
pushd ${RELEASE_DIR}
./configure
make

popd
popd
mkdir -p ${TARGET_DIR}
cp /tmp/${RELEASE_DIR}/daemonize ${TARGET_DIR}
rm -rf /tmp/${RELEASE_DIR}
