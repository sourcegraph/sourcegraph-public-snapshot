#!/bin/sh

# This script installs ctags within an alpine container.

# Commit hash of github.com/universal-ctags/ctags.
# Last bumped 2022-02-28
# This version includes a fix that hasn't landed on master yet:
# https://github.com/universal-ctags/ctags/pull/3300
CTAGS_VERSION=90a16c009c52a35578140c6c731bcd5faa104f11

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
