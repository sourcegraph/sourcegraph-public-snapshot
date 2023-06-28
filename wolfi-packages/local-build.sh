#!/usr/bin/env bash

set -eu -o pipefail

# This script can be used to quickly build packages locally when working on package configs in this directory.
# In production, packages are built using the CI pipeline.

cd "$(dirname "${BASH_SOURCE[0]}")/../"
cd "wolfi-packages"

PACKAGE_DIR="local-repo/packages"
KEY_DIR="local-repo/keys"

mkdir -p "$PACKAGE_DIR" "$KEY_DIR"

# Generate keys for local repository
if [ ! -f "local-repo/keys/melange.rsa" ]; then
  echo " 🗝️  Initializing keypair for local repo..."
  docker run \
    -v "$PWD/local-repo/keys/":/keys \
    cgr.dev/chainguard/melange keygen "/keys/melange.rsa"

  if [ -f "local-repo/keys/melange.rsa" ]; then
    echo " 🔐 Keypair initialized"
  else
    echo " ❗️ Error initializing keypair"
    exit 1
  fi
fi

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
  -v "$PWD/$PACKAGE_DIR":/work/packages \
  -v "$PWD/$KEY_DIR":/keys \
  cgr.dev/chainguard/melange build "$file_name" --arch x86_64

echo " ✅  Built package '$package_name' under '$PACKAGE_DIR'"
echo " 🐳  Use in locally-built base images with '${package_name}@local'"

# TODO: Sign packages
# TODO: Preserve melange tmp dir on failure
