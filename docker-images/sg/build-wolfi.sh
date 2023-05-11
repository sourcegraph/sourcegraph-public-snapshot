#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

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

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/dev/sg"
go build -trimpath -ldflags "-X main.BuildCommit=$BUILD_COMMIT" -o "$OUTPUT/sg" -buildmode exe "$pkg"

echo "--- docker build $IMAGE"
docker build -f docker-images/sg/Dockerfile.wolfi -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
