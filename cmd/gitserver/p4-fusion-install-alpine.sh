#!/bin/sh

# This script installs p4-fusion within an alpine container.

set -eu

tmpdir=$(mktemp -d)
cd "$tmpdir"

cleanup() {
  echo "--- cleanup"
  apk --no-cache --purge del p4-build-deps 2>/dev/null || true
  cd /
  rm -rf "$tmpdir" || true
}

trap cleanup EXIT

test_p4_fusion() {
  # Test that p4-fusion runs and is on the path
  echo "--- p4-fusion test"
  ldd "$(which p4-fusion)"
  p4-fusion >/dev/null
}

set -x

# Hello future traveler. Building p4-fusion is one of our slowest steps in CI.
# Luckily the versions very rarely change and nearly everything is statically
# linked. This means we can manually upload the output of this build script to
# a bucket and save lots of time.
#
# If the version has changed please add it to the sha256sum in the prebuilt
# binary check. You can run
#
#   docker build -t p4-fusion --target=p4-fusion .
#
# Then extract the binary from /usr/local/bin/p4-fusion. Please rename it
# follow the format and upload to the bucket here
# https://console.cloud.google.com/storage/browser/sourcegraph-artifacts/p4-fusion
export P4_FUSION_VERSION=v1.11

# Runtime dependencies
echo "--- p4-fusion apk runtime-deps"
apk add --no-cache libstdc++

# Check if we have a prebuilt binary
echo "--- p4-fusion prebuilt binary check"
if wget https://storage.googleapis.com/sourcegraph-artifacts/p4-fusion/p4-fusion-"$P4_FUSION_VERSION"-musl-x86_64; then
  src=p4-fusion-"$P4_FUSION_VERSION"-musl-x86_64
  cat <<EOF | grep "$src" | sha256sum -c
98c4991b40cdd0e0cad2fd5fdbe9f8ff56901415dd1684aed5b4531fc49ab79e  p4-fusion-v1.11-musl-x86_64
EOF
  chmod +x "$src"
  mv "$src" /usr/local/bin/p4-fusion
  test_p4_fusion
  exit 0
fi

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

# Fetching p4 sources archive
echo "--- p4-fusion fetch"
mkdir p4-fusion-src
wget https://github.com/salesforce/p4-fusion/archive/refs/tags/"$P4_FUSION_VERSION".tar.gz
tar -C p4-fusion-src -xzf "$P4_FUSION_VERSION".tar.gz --strip 1

# It should be possible to build against the latest 1.x version of OpenSSL.
# However, Perforce recommends linking against the same minor version of
# OpenSSL that is referenced in the Helix Core C++ API for best compatibility.
# https://www.perforce.com/manuals/p4api/Content/P4API/client.programming.compiling.html#SSL_support
echo "--- p4-fusion openssl fetch"
mkdir openssl-src
wget https://www.openssl.org/source/openssl-1.0.2t.tar.gz
tar -C openssl-src -xzf openssl-1.0.2t.tar.gz --strip 1

echo "--- p4-fusion openssl build"
cd openssl-src
./config
# We only need libcrypto and libssl, which "build_libs" covers. Note: using
# unbounded concurrency caused flakes on CI.
make build_libs

echo "--- p4-fusion openssl install"
# TODO "install" includes "all". Can we avoid extra work?
make install
cd ..

# We also need Helix Core C++ API to build p4-fusion
echo "--- p4-fusion helix-core fetch"
mkdir -p p4-fusion-src/vendor/helix-core-api/linux
wget https://www.perforce.com/downloads/perforce/r22.1/bin.linux26x86_64/p4api.tgz
tar -C p4-fusion-src/vendor/helix-core-api/linux -xzf p4api.tgz --strip 1

# Build p4-fusion
echo "--- p4-fusion build"
cd p4-fusion-src
./generate_cache.sh RelWithDebInfo
./build.sh
cd ..

# Move exe file to /usr/local/bin where other executables are located
echo "--- p4-fusion install"
mv p4-fusion-src/build/p4-fusion/p4-fusion /usr/local/bin

test_p4_fusion
