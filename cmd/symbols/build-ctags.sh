#!/usr/bin/env bash

# This script builds the ctags image for local development.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

# If CTAGS_COMMAND is set, we don't need to build the image. See
# ./universal-ctags-dev.
if [[ -n "${CTAGS_COMMAND}" ]]; then
  echo "CTAGS_COMMAND set. Building of Docker image not necessary."
  exit 0
fi

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

cp -a ./cmd/symbols/.ctags.d "$OUTPUT"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"

# Build ctags docker image for universal-ctags-dev
echo "Building ctags docker image"
docker build -f cmd/symbols/Dockerfile -t ctags "$OUTPUT" \
  --target=ctags \
  --progress=plain \
  --quiet >/dev/null
