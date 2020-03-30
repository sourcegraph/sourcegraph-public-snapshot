#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eux

OUTPUT=`mktemp -d -t sgdockerbuild_XXXXXXX`
cleanup() {
    rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

cp -a ./cmd/symbols/.ctags.d "$OUTPUT"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"
cp -a ./dev/libsqlite3-pcre/install-alpine.sh "$OUTPUT/libsqlite3-pcre-install-alpine.sh"

export OUTPUT
./cmd/symbols/go-build.sh

# Build either symbols or ctags (for dev)
DOCKER_TARGET="${DOCKER_TARGET:-symbols}"

echo "--- docker build"
docker build -f cmd/symbols/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --target="${DOCKER_TARGET}" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
