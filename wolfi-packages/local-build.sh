#!/usr/bin/env bash

set -eu -o pipefail

# This script can be used to quickly build packages locally when working on package configs in this directory.
# In production, packages are built using the CI pipeline.

cd "$(dirname "${BASH_SOURCE[0]}")/../"
cd "wolfi-packages"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build e.g. ./local-build.sh coursier.yaml"
  exit 0
fi

# Normalise name by adding .yaml if necessary
name=${1%/}
name=$(echo "$name" | sed -r 's/^([a-zA-Z0-9_-]+)$/\1.yaml/')

if [ ! -f "$name" ]; then
  echo "File '$name' does not exist"
  exit 1
fi

echo "Building package '$name'"

# Mounting /tmp can be useful for debugging: -v "$HOME/tmp":/tmp \
docker run --privileged \
  -v "$PWD":/work \
  cgr.dev/chainguard/melange build "$name" --arch x86_64
