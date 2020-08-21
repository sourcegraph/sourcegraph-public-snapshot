#!/usr/bin/env bash

# This script builds the ctags image for local development.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

cp -a ./cmd/symbols/.ctags.d "$OUTPUT"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"

# Build ctags docker image for universal-ctags-dev
echo "--- Building ctags docker image"
docker build -f cmd/symbols/Dockerfile -t ctags "$OUTPUT" \
  --target=ctags \
  --progress=plain \
  --quiet >/dev/null
