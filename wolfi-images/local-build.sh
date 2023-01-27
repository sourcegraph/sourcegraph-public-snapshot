#!/usr/bin/env bash

set -euo pipefail

# This script can be used to quickly build base images locally when working on image configs in this directory.
# In production, base images are built using the CI pipeline.

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the image name to build e.g. './local-build.sh sourcegraph'"
  exit 1
fi

name=${1%/}

if [ ! -f "$name.yaml" ]; then
  echo "File '$name.yaml' does not exist"
  exit 1
fi

## Build base image using apko build container
echo " * Building base image '$name' using apko"
docker run \
  -v "$PWD":/work \
  cgr.dev/chainguard/apko \
  build --debug "${name}.yaml" \
  "sourcegraph-wolfi/$name-base:latest" \
  "sourcegraph-wolfi-$name-base.tar" ||
  (echo "*** Build failed ***" && exit 1)

# To build images against a local repo with a custom signing key:
# Pass volumes to Docker:
#   -v "$PWD/../wolfi-packages/packages":/work/packages \
#   -v "$PWD/../wolfi-packages/keys":/work/keys \
# Pass signing key to apko:
#   -k /work/keys/melange.rsa.pub

## Import into Docker
echo " * Loading tarball into Docker"
docker load <"sourcegraph-wolfi-$name-base.tar"

## Cleanup
echo " * Cleaning up tarball and SBOM"
rm "sourcegraph-wolfi-$name-base.tar"
rm sbom*
rmdir keys/ packages/
