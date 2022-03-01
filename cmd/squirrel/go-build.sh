#!/usr/bin/env bash

# This script builds the squirrel go binary.
# Requires a single argument which is the path to the target bindir.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT="${1:?no output path provided}"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux

# go-tree-sitter depends on cgo. Without cgo, it will build but it'll throw an error at query time.
export CGO_ENABLED=1

# Default CC to musl-gcc.
export CC="${CC:-musl-gcc}"

help() {
  echo "You need to set CC to a musl compiler in order to compile go-tree-sitter for Alpine."
  echo
  echo "    Linux: run 'apt-get install -y musl-tools'"
  echo "    macOS: download https://github.com/FiloSottile/homebrew-musl-cross/blob/6ee3329ee41231fe693306490f8e4d127c70e618/musl-cross.rb and run 'brew install ~/Downloads/musl-cross.rb'"
}

if ! command -v "$CC" >/dev/null; then
  echo "$CC not found."
  help
  exit 1
fi

# Make sure this is a musl compiler.
case "$CC" in
  *musl*)
    ;;
  *)
    echo "$CC doesn't look like a musl compiler."
    help
    exit 1
    ;;
esac

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/cmd/squirrel"
env go build \
  -trimpath \
  -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
  -buildmode exe \
  -tags dist \
  -o "$OUTPUT/$(basename $pkg)" \
  "$pkg"
