#!/bin/sh

# This script installs universal-ctags within an alpine container.

# Commit hash of github.com/universal-ctags/ctags.
# Last bumped 2023-10-24.
# When bumping please remember to also update Zoekt: https://github.com/sourcegraph/zoekt/blob/d3a8fbd8385f0201dd54ab24114ebd588dfcf0d8/install-ctags-alpine.sh
CTAGS_VERSION=v6.0.0
CTAGS_ARCHIVE_TOP_LEVEL_DIR=ctags-6.0.0
# When using commits you can rely on
# CTAGS_ARCHIVE_TOP_LEVEL_DIR=ctags-$CTAGS_VERSION

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
curl --retry 5 "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" | tar xz -C /tmp
cd /tmp/$CTAGS_ARCHIVE_TOP_LEVEL_DIR
./autogen.sh
./configure --program-prefix=universal- --enable-json
make -j"$NUMCPUS" --load-average="$NUMCPUS"
make install
