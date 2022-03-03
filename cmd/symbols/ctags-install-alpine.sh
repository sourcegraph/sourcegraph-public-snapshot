#!/bin/sh

# This script installs ctags within an alpine container.

# Commit hash of github.com/universal-ctags/ctags.
# Last bumped 2022-02-10
CTAGS_VERSION=37a4b3601288bcdc02a387197ff8d9b971f7ab34

cleanup() {
  apk --no-cache --purge del ctags-build-deps || true
  cd /
  rm -rf /tmp/ctags-$CTAGS_VERSION
}

trap cleanup EXIT

set -eux

apk --no-cache add \
  --virtual ctags-build-deps \
  autoconf \
  automake \
  binutils \
  curl \
  g++ \
  gcc \
  jansson-dev \
  make \
  pkgconfig

# ctags is dynamically linked against jansson
apk --no-cache add jansson

NUMCPUS=$(grep -c '^processor' /proc/cpuinfo)

# Installation
curl "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" | tar xz -C /tmp
cd /tmp/ctags-$CTAGS_VERSION
./autogen.sh
./configure --program-prefix=universal- --enable-json
make -j"$NUMCPUS" --load-average="$NUMCPUS"
make install
