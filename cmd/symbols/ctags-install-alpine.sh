#!/bin/sh

# This script installs ctags within an alpine container.

set -eux

# Commit hash of github.com/universal-ctags/ctags
CTAGS_VERSION=681a8d5f5f6fcca9d8cca703c250f4cdf05b45c3

apk --no-cache add \
  autoconf \
  automake \
  binutils \
  curl \
  g++ \
  gcc \
  jansson-dev \
  libseccomp-dev \
  jansson \
  libseccomp \
  linux-headers \
  make \
  pkgconfig

# Installation
curl "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" | tar xz -C /tmp
cd /tmp/ctags-$CTAGS_VERSION
./autogen.sh
./configure --program-prefix=universal- --enable-json --enable-seccomp
make -j8
make install

# Cleanup
cd /
rm -rf /tmp/ctags-$CTAGS_VERSION
