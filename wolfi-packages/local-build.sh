#!/usr/bin/env bash

set -eu -o pipefail

# This script can be used to quickly build packages locally when working on package configs in this directory.
# In production, packages are built using the CI pipeline.

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build"
  exit 0
fi

name=${1%/}
echo "Building package '$name'"

# Mounting /tmp can be useful for debugging: -v "$HOME/tmp":/tmp \
docker run --privileged \
  -v "$PWD":/work \
  cgr.dev/chainguard/melange build "$name" --arch x86_64
