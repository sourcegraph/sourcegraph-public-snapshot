#!/usr/bin/env bash

set -euo pipefail

OUTPUT=$(mktemp -d -t sghashbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT" 2>/dev/null
}
trap cleanup EXIT

# Do not embed build flags to produce a stable output
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/executor"
artifact="$OUTPUT/$(basename $pkg)"
go build -trimpath -buildmode exe -tags dist -o "$artifact" "$pkg"

# Generate hash for build artifact
md5sum <"$artifact"

# Generate hash for entire build directory
tar c --exclude='./enterprise/cmd/executor/docker-mirror' ./enterprise/cmd/executor | md5sum
