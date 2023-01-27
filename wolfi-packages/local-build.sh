#!/usr/bin/env bash

set -eu -o pipefail

# This script can be used to quickly build packages locally when working on package configs in this directory.
# In production, packages are built using the CI pipeline.

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build"
  exit 0
fi

docker run --privileged \
  -v "$PWD":/work
# -v "$HOME/tmp":/tmp \ # Useful for debugging
cgr.dev/chainguard/melange build "$1" --arch x86_64
