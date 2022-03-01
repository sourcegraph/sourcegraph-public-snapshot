#!/usr/bin/env bash

# This script builds the squirrel docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Build go binary into $OUTPUT
./cmd/squirrel/go-build.sh "$OUTPUT"

echo "--- docker build"
docker build -f cmd/squirrel/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
