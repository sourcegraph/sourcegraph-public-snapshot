#!/bin/sh

# This script installs ctags within an alpine container.

# Commit hash of github.com/universal-ctags/ctags
CTAGS_VERSION=7c4df9d38c4fe4bb494e5f3b2279034d7d8bd7b7

cleanup() {
  cd /
  rm -rf /tmp/ctags-$CTAGS_VERSION
}

trap cleanup EXIT

set -eux

apk --no-cache add \
  autoconf \
  automake \
  binutils \
  curl \
  g++ \
  gcc \
  jansson-dev \
  jansson \
  make \
  pkgconfig

NUMCPUS=$(grep -c '^processor' /proc/cpuinfo)

# Installation
curl "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" | tar xz -C /tmp
cd /tmp/ctags-$CTAGS_VERSION
./autogen.sh
./configure --program-prefix=universal- --enable-json
make -j"$NUMCPUS" --load-average="$NUMCPUS"
make install
