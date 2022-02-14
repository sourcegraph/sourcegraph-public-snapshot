#!/bin/sh

# This script installs p4-fusion within an alpine container.

cleanup() {
  apk --no-cache --purge del p4-build-deps || true
  cd /
  rm -rf /usr/local/bin/p4-fusion-src || true
  rm -rf /usr/local/bin/v1.5.tar.gz || true
}

trap cleanup EXIT

cd /

set -eux

# Runtime dependencies
apk add --no-cache libstdc++

# Build dependencies
apk add --no-cache \
  --virtual p4-build-deps \
  wget \
  g++ \
  gcc \
  perl \
  bash \
  cmake \
  make

# Fetching p4 sources archive
wget https://github.com/salesforce/p4-fusion/archive/refs/tags/v1.5.tar.gz
mv v1.5.tar.gz /usr/local/bin
mkdir -p /usr/local/bin/p4-fusion-src
tar -C /usr/local/bin/p4-fusion-src -xzvf /usr/local/bin/v1.5.tar.gz --strip 1

# We need a specific version of OpenSSL
wget https://www.openssl.org/source/openssl-1.0.2t.tar.gz
tar -xzvf openssl-1.0.2t.tar.gz
cd /openssl-1.0.2t

./config && make && make install

# We also need Helix Core C++ API to build p4-fusion
wget https://www.perforce.com/downloads/perforce/r21.1/bin.linux26x86_64/p4api.tgz
mkdir -p /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux
mv p4api.tgz /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux
tar -C /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux -xzvf /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux/p4api.tgz --strip 1

cd /usr/local/bin/p4-fusion-src

# Build p4-fusion
./generate_cache.sh Release
./build.sh

# Move exe file to /usr/local/bin where other executables are located
mv /usr/local/bin/p4-fusion-src/build/p4-fusion/p4-fusion /usr/local/bin
