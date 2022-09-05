#!/usr/bin/env bash

# This script builds the symbols go binary.
# Requires a single argument which is the path to the target bindir.
#
# To test you can run
#
#   VERSION=test ./cmd/symbols/go-build.sh /tmp

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT="${1:?no output path provided}"

echo "--- docker symbols build"

# Required due to use of RUN --mount=type=cache in Dockerfile.
export DOCKER_BUILDKIT=1

docker build -f cmd/symbols/Dockerfile -t symbols-build "$(pwd)" \
  --target=symbols-build \
  --progress=plain \
  --build-arg VERSION \
  --build-arg PKG="${PKG:-github.com/sourcegraph/sourcegraph/cmd/symbols}"

docker cp "$(docker create --rm symbols-build)":/symbols "$OUTPUT/symbols"
