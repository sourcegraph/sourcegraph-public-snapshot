#!/usr/bin/env bash

set -euo pipefail

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the apko image name to build e.g. 'sourcegraph-wolfi'"
  exit 1
fi

name=${1%/}

if [ ! -d "$name" ]; then
  echo "Directory '$name' does not exist"
  exit 1
fi

if [ ! -f "$name/apko.yaml" ]; then
  echo "File '$name/apko.yaml' does not exist"
  exit 1
fi

cd "$name"

## Build base image using apko build container
echo " * Building apko base image '$name'"
docker run \
  -v "$PWD":/work \
  -v "$PWD/../../dependencies/packages":/work/packages \
  -v "$PWD/../../dependencies/keys":/work/keys \
  cgr.dev/chainguard/apko \
  build --debug -k /work/keys/melange.rsa.pub apko.yaml \
  "sourcegraph-wolfi/$name-base:latest" \
  "sourcegraph-wolfi-$name-base.tar" ||
  (echo "*** Build failed ***" && exit 1)

## Import into Docker
echo " * Loading tarball into Docker"
docker load <"sourcegraph-wolfi-$name-base.tar"

## Cleanup
echo " * Cleaning up tarball and SBOM"
rm "sourcegraph-wolfi-$name-base.tar"
rm sbom*
rmdir keys/ packages/
