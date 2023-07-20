#!/usr/bin/env bash

# This script builds the ctags images for local development.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

# If CTAGS_COMMAND is set to a custom executable, we don't need to build the
# image. See /dev/universal-ctags-dev.
if [[ "${CTAGS_COMMAND}" == "dev/universal-ctags-dev" ]]; then
  echo "CTAGS_COMMAND set to custom executable. Building of Docker image not necessary."
else
  # Build ctags docker image for universal-ctags-dev
  echo "Building universal-ctags docker image"
  docker build -f cmd/symbols/Dockerfile -t ctags . \
    --platform linux/amd64 \
    --target=ctags \
    --progress=plain
fi

# If SCIP_CTAGS_COMMAND is set to a custom executable, we don't need to build the
# image. See /dev/scip-ctags-dev.
if [[ "${SCIP_CTAGS_COMMAND}" != "dev/scip-ctags-dev" ]]; then
  echo "SCIP_CTAGS_COMMAND set to custom executable. Building of Docker image or Rust code not necessary."
else
  if [[ "$(uname -m)" == "arm64" ]]; then
    # build ctags with cargo; prevent x86-64 slowdown on mac
    root="$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null
    "$root"/dev/scip-ctags-install.sh
  else
    # Build ctags docker image for scip-ctags-dev
    echo "Building scip-ctags docker image"
    docker build -f cmd/symbols/Dockerfile -t scip-ctags . \
      --platform linux/amd64 \
      --target=scip-ctags \
      --progress=plain
  fi
fi
