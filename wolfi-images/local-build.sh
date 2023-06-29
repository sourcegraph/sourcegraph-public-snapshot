#!/usr/bin/env bash

set -euo pipefail

# This script can be used to quickly build base images locally when working on image configs in this directory.
# In production, base images are built using the CI pipeline.

cd "$(dirname "${BASH_SOURCE[0]}")/../"
BASE_DIR=$(pwd)

PACKAGE_DIR="$BASE_DIR/wolfi-packages/local-repo/packages"
KEY_DIR="$BASE_DIR/wolfi-packages/local-repo/keys"
KEY_FILENAME="sourcegraph-dev-local.rsa"
KEY_FILEPATH="$KEY_DIR/$KEY_FILENAME"

mkdir -p "$PACKAGE_DIR" "$KEY_DIR"

# Generate keys for local repository
if [ ! -f "$KEY_FILEPATH" ]; then
  echo " 🗝️  Initializing keypair for local repo..."
  docker run \
    -v "$KEY_DIR":/keys \
    cgr.dev/chainguard/melange keygen "/keys/$KEY_FILENAME"

  if [ -f "$KEY_FILEPATH" ]; then
    echo " 🔐 Keypair initialized"
  else
    echo " ❗️ Error initializing keypair"
    exit 1
  fi
fi

cd "wolfi-images"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the image name to build e.g. './local-build.sh sourcegraph'"
  exit 1
fi

# Normalise name
image_name=${1%.yaml}
file_name="${image_name}.yaml"

if [ ! -f "$file_name" ]; then
  echo "File '$file_name' does not exist"
  exit 1
fi

## Build base image using apko build container
echo " 📦 Building base image '$image_name' using apko..."
docker run \
  -v "$PWD":/work \
  -v "$KEY_DIR:/keys" \
  -v "$PACKAGE_DIR:/packages" \
  -e SOURCE_DATE_EPOCH="$(date +%s)" \
  cgr.dev/chainguard/apko \
  build \
  --debug "${file_name}" \
  --arch x86_64 \
  --repository-append "@local /packages" --keyring-append "/keys/$KEY_FILENAME.pub" \
  "sourcegraph-wolfi/$image_name-base:latest" \
  "sourcegraph-wolfi-$image_name-base.tar" ||
  (echo "*** Build failed ***" && exit 1)

## Import into Docker
echo " * Loading tarball into Docker"
docker load <"sourcegraph-wolfi-$image_name-base.tar"

## Cleanup
echo " * Cleaning up tarball and SBOM"
rm "sourcegraph-wolfi-$image_name-base.tar"
rm sbom*
