#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

set -euf -o pipefail
tmpdir=$(mktemp -d -t melange-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  rm -rf "$tmpdir"
}
trap cleanup EXIT

(
  cd "$tmpdir"
  mkdir bin

  # Install melange
  wget https://github.com/chainguard-dev/melange/releases/download/v0.2.0/melange_0.2.0_linux_amd64.tar.gz
  tar zxf melange_0.2.0_linux_amd64.tar.gz
  mv melange_0.2.0_linux_amd64/melange bin/melange

  # Install apk
  wget https://gitlab.alpinelinux.org/alpine/apk-tools/-/package_files/62/download -O bin/apk
  chmod +x bin/apk

  # Fetch custom-built bubblewrap 0.7.0 (temporary, until https://github.com/sourcegraph/infrastructure/pull/4520 is merged)
  wget https://dollman.org/files/bwrap
  chmod +x bwrap
  mv bwrap bin/
)

export PATH="$tmpdir/bin:$PATH"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build"
  exit 0
fi

name=${1%/}

cd "wolfi-packages"

if [ ! -e "${name}.yaml" ]; then
  echo "File '$name.yaml' does not exist"
  exit 1
fi

# NOTE: Melange relies upon a more recent version of bubblewrap than ships with Ubuntu 20.04. We therefore build a recent
# bubblewrap release in buildkite-agent-stateless-bazel's Dockerfile, and ship it in /usr/local/bin

echo " * Building melange package '$name'"
# TODO: Signing key
melange build "$name.yaml" --arch x86_64

# Upload package as build artifact
buildkite-agent artifact upload packages/*/*
