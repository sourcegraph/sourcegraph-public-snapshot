#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eux

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

cp -a ./cmd/precise-code-intel "$OUTPUT"

echo "--- docker build"
docker build -f cmd/precise-code-intel/bundle-manager/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
