#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT=`mktemp -d -t sgdockerbuild_XXXXXXX`
cleanup() {
    rm -rf "$OUTPUT"
}
trap cleanup EXIT


# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export OUTPUT # build artifact goes here
./cmd/symbols/go-build.sh

cp -a ./cmd/symbols/.ctags.d "$OUTPUT"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"
cp -a ./dev/libsqlite3-pcre/install-alpine.sh "$OUTPUT/libsqlite3-pcre-install-alpine.sh"

echo "--- docker build"
docker build -f cmd/symbols/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --target="${DOCKER_TARGET:-symbols}" \
    ${DOCKER_BUILD_FLAGS:-} \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
