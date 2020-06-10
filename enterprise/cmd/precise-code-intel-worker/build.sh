#!/usr/bin/env bash

# This script builds the precise-code-intel-worker docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

cp -a ./dev/libsqlite3-pcre/install-alpine.sh "$OUTPUT/libsqlite3-pcre-install-alpine.sh"

# Build go binary into $OUTPUT
./enterprise/cmd/precise-code-intel-worker/go-build.sh "$OUTPUT"

echo "--- docker build"
docker build -f enterprise/cmd/precise-code-intel-worker/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
