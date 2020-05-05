#!/usr/bin/env bash

# This script builds the src-expose docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)

cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

# Entrypoint script and src-expose binary
cp -a ./dev/src-expose/entry.sh "$OUTPUT"
go build -trimpath -o "$OUTPUT/src-expose" github.com/sourcegraph/sourcegraph/dev/src-expose

docker build -f dev/src-expose/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
