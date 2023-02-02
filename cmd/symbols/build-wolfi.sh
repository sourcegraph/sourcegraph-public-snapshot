#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

echo "--- docker build symbols"
docker build -f cmd/symbols/Dockerfile.wolfi -t "$IMAGE" "$(pwd)" \
  --platform="${PLATFORM:-linux/amd64}" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION \
  --build-arg PKG="${PKG:-github.com/sourcegraph/sourcegraph/cmd/symbols}"
