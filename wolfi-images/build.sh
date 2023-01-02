#!/usr/bin/env bash

set -euo pipefail

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the apko image name to build e.g. 'sourcegraph-wolfi'"
  exit 1
fi

if [ ! -d "$1" ]; then
  echo "Directory '$1' does not exist"
  exit 1
fi

if [ ! -f "$1/apko.yaml" ]; then
  echo "File '$1/apko.yaml' does not exist"
  exit 1
fi

cd "$1"

echo "Building apko base image '$1'"

# # Build base image using apko build container
docker run \
  -v "$PWD":/work \
  -v "$PWD/../../dependencies/packages":/work/packages \
  -v "$PWD/../../dependencies/keys":/work/keys \
  cgr.dev/chainguard/apko \
  build --debug -k /work/keys/melange.rsa.pub apko.yaml \
  "sourcegraph/$1:latest" \
  "sourcegraph-$1.tar" ||
  echo "*** Build failed ***"

# # Import into Docker
docker load <"sourcegraph-$1".tar
