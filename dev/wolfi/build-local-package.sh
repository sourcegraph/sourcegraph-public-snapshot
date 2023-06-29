#!/usr/bin/env bash

set -eu -o pipefail

# This script can be used to quickly build packages locally when working on package configs in this directory.
# In production, packages are built using the CI pipeline.

cd "$(dirname "${BASH_SOURCE[0]}")/../"
BASE_DIR=$(pwd)

PACKAGE_DIR="$BASE_DIR/wolfi-packages/local-repo/packages"
ARCH="x86_64"
KEY_DIR="$BASE_DIR/wolfi-packages/local-repo/keys"
KEY_FILENAME="sourcegraph-dev-local.rsa"
KEY_FILE="$KEY_DIR/$KEY_FILENAME"

mkdir -p "$PACKAGE_DIR/$ARCH" "$KEY_DIR"

# Generate keys for local repository
if [ ! -f "$KEY_FILE" ]; then
  echo " 🗝️  Initializing keypair for local repo..."
  docker run \
    -v "$KEY_DIR":/keys \
    cgr.dev/chainguard/melange keygen "/keys/$KEY_FILENAME"

  if [ -f "$KEY_FILE" ]; then
    echo " 🔐 Keypair initialized"
  else
    echo " ❗️ Error initializing keypair"
    exit 1
  fi
fi

cd "wolfi-packages"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build e.g. ./local-build.sh coursier.yaml"
  exit 0
fi

# Get first variable and strip off the .yaml suffix, if it exists
package_name=${1%.yaml}
file_name="${package_name}.yaml"

echo " 📦 Building package '$package_name'..."

if [ ! -f "$file_name" ]; then
  echo " ❌  Package manifest file '$file_name' does not exist"
  exit 1
fi

# Create a temporary directory
tmpdir=$(mktemp -d -t melange-build.XXXXXXXX)
# trap 'rm -r $tmpdir' EXIT

# Copy package file + folder (if present) to a temporary directory
cp "$file_name" "$tmpdir"
if [ -d "$package_name" ]; then
  cp -r "$package_name" "$tmpdir"
fi

# Mounting /tmp can be useful for debugging: -v "$HOME/tmp":/tmp \
docker run --privileged \
  -v "$tmpdir":/work \
  -v "$PACKAGE_DIR":/work/packages \
  -v "$KEY_DIR":/keys \
  cgr.dev/chainguard/melange \
  build "$file_name" \
  --arch x86_64 \
  --signing-key "/keys/$KEY_FILENAME"

echo -e "\n"
echo " ✅  Built package '$package_name' under '$PACKAGE_DIR'"
echo " 🐳  Use this package in locally-built base images by adding the package '${package_name}@local'"

# TODO: Preserve melange tmp dir on failure
