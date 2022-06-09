#!/bin/sh

# This script installs ctags within an alpine container.

# Commit hash of github.com/universal-ctags/ctags.
# Last bumped 2022-04-04.
# When bumping please remember to also update Zoekt: https://github.com/sourcegraph/zoekt/blob/d3a8fbd8385f0201dd54ab24114ebd588dfcf0d8/install-ctags-alpine.sh
CTAGS_VERSION=f95bb3497f53748c2b6afc7f298cff218103ab90

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
