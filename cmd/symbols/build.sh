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

# Get additional build args
. ./cmd/symbols/go-build-args.sh

echo "--- go build"
for pkg in github.com/sourcegraph/sourcegraph/cmd/symbols; do
    go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION" -buildmode exe -tags dist -o $OUTPUT/$(basename $pkg) $pkg
done

cp -a ./cmd/symbols/.ctags.d "$OUTPUT"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"
cp -a ./dev/libsqlite3-pcre/install-alpine.sh "$OUTPUT/libsqlite3-pcre-install-alpine.sh"

echo "--- docker build"
docker build -f cmd/symbols/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
