#!/bin/sh

# This script installs libsqlite3-pcre within an alpine container.

set -eux

# Commit hash of github.com/ralight/sqlite3-pcre
SQLITE3_PCRE_VERSION=c98da412b431edb4db22d3245c99e6c198d49f7a

apk --no-cache add \
  --virtual build-deps \
  curl \
  gcc \
  git \
  libc-dev \
  make \
  pcre-dev \
  sqlite-dev

# Installation
mkdir /sqlite3-pcre
curl -fsSL "https://codeload.github.com/ralight/sqlite3-pcre/tar.gz/$SQLITE3_PCRE_VERSION" | tar -C /sqlite3-pcre -xzvf - --strip 1
cd /sqlite3-pcre
make

# Cleanup
apk --no-cache --purge del build-deps
