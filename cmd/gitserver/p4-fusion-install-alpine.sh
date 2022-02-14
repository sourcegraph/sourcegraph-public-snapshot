#!/bin/sh

# This script installs p4-fusion within an alpine container.

set -eu

tmpdir=$(mktemp -d)

cleanup() {
  echo "--- cleanup"
  apk --no-cache --purge del p4-build-deps || true
  cd /
  rm -rf "$tmpdir" || true
}

trap cleanup EXIT

set -x

# Runtime dependencies
echo "--- p4-fusion apk runtime-deps"
apk add --no-cache libstdc++

# Build dependencies
echo "--- p4-fusion apk build-deps"
apk add --no-cache \
  --virtual p4-build-deps \
  wget \
  g++ \
  gcc \
  perl \
  bash \
  cmake \
  make

cd "$tmpdir"

# Fetching p4 sources archive
echo "--- p4-fusion fetch"
mkdir p4-fusion-src
wget https://github.com/salesforce/p4-fusion/archive/refs/tags/v1.5.tar.gz
tar -C p4-fusion-src -xzf v1.5.tar.gz --strip 1

# We need a specific version of OpenSSL
echo "--- p4-fusion openssl fetch"
mkdir openssl-src
wget https://www.openssl.org/source/openssl-1.0.2t.tar.gz
tar -C openssl-src -xzf openssl-1.0.2t.tar.gz --strip 1

echo "--- p4-fusion openssl build"
cd openssl-src
./config
# We only need libcrypto and libssl, which "build_libs" covers. We use
# unbounded concurrency. Experiments on a 32-core machine showed a multiple of
# 32 didn't help.
make -j build_libs

echo "--- p4-fusion openssl install"
# TODO "install" includes "all". Can we avoid extra work?
make install
cd ..

# We also need Helix Core C++ API to build p4-fusion
echo "--- p4-fusion helix-core fetch"
mkdir -p p4-fusion-src/vendor/helix-core-api/linux
wget https://www.perforce.com/downloads/perforce/r21.1/bin.linux26x86_64/p4api.tgz
tar -C p4-fusion-src/vendor/helix-core-api/linux -xzf p4api.tgz --strip 1

# Build p4-fusion
echo "--- p4-fusion build"
cd p4-fusion-src
./generate_cache.sh Release
./build.sh
cd ..

# Move exe file to /usr/local/bin where other executables are located
echo "--- p4-fusion install"
mv p4-fusion-src/build/p4-fusion/p4-fusion /usr/local/bin

# Test that p4-fusion runs and is on the path
echo "--- p4-fusion test"
ldd "$(which p4-fusion)"
p4-fusion >/dev/null
