#!/bin/bash

set -euf -o pipefail
tmpdir=$(mktemp -d -t wolfi-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  rm -rf "$tmpdir"
}
trap cleanup EXIT

(
  cd $tmpdir

  # Install apko
  wget https://github.com/chainguard-dev/apko/releases/download/v0.6.0/apko_0.6.0_linux_amd64.tar.gz
  tar zxf apko_0.6.0_linux_amd64.tar.gz
  mv apko_0.6.0_linux_amd64/apko bin/apko

  # Install apk
  mkdir bin
  wget https://gitlab.alpinelinux.org/alpine/apk-tools/-/package_files/62/download -O bin/apk
  chmod +x bin/apk
)

export PATH="$tmpdir/bin:$PATH"

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

echo " * Building apko base image '$name'"
apko build --debug apko.yaml \
  "sourcegraph-wolfi/${name}-base:latest" \
  "sourcegraph-wolfi-${name}-base.tar" ||
  (echo "*** Build failed ***" && exit 1)
