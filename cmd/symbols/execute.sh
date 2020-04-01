#!/usr/bin/env bash

# This script builds the ctags image and the symbols go binary, then runs the go binary.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

# Build ctags docker image for universal-ctags-dev
IMAGE=ctags DOCKER_TARGET=ctags DOCKER_BUILD_FLAGS="--quiet" ./cmd/symbols/build.sh

# Build and run symbols binary
./dev/libsqlite3-pcre/build.sh
OUTPUT=./.bin ./cmd/symbols/go-build.sh
LIBSQLITE3_PCRE="$(./dev/libsqlite3-pcre/build.sh libpath)" ./.bin/symbols
