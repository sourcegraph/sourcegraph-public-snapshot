#!/usr/bin/env bash

set -euo pipefail

# This script calculates a hash of the executor binary and the metadata files.
# This is used in CI to determine if there was any change made since the last

OUTPUT=$(mktemp -d -t sghashbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT" 2>/dev/null
}
trap cleanup EXIT

# Do not embed build flags to produce a stable output.
# Ensure a stable environment for building the binary.
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0
pkg="github.com/sourcegraph/sourcegraph/cmd/executor"
artifact="$OUTPUT/$(basename $pkg)"
go build -buildvcs=false -trimpath -buildmode exe -tags dist -o "$artifact" "$pkg"

# Generate hash for build artifact
md5sum <"$artifact"

# Generate hash for entire build directory
tar c \
  --exclude='./cmd/executor/docker-mirror' \
  --exclude='./cmd/executor/docker-image' \
  --exclude='./cmd/executor/kubernetes' \
  ./cmd/executor | md5sum
