#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

echo "--- docker build symbols"
docker build -f cmd/symbols/Dockerfile -t "$IMAGE" "$(pwd)" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION \
  --build-arg PKG="${PKG:-github.com/sourcegraph/sourcegraph/cmd/symbols}"
