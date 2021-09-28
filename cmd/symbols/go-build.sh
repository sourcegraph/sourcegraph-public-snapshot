#!/usr/bin/env bash

# This script builds the symbols go binary.
# Requires a single argument which is the path to the target bindir.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT="${1:?no output path provided}"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux

# go-sqlite3 depends on cgo. Without cgo, it will build but it'll throw an error at query time.
export CGO_ENABLED=1

# Ensure musl-gcc is available since we're building to run on Alpine, which uses musl.
if ! command -v musl-gcc >/dev/null; then
  echo "musl-gcc not found, which is needed for cgo for go-sqlite3. Run 'apt-get install -y musl-tools'."
  exit 1
fi

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/cmd/symbols"
env CC=musl-gcc go build \
  -trimpath \
  -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
  -buildmode exe \
  -tags dist \
  -o "$OUTPUT/$(basename $pkg)" \
  "$pkg"

# Make sure go-sqlite3 was compiled with cgo.
echo "--- sanity check"
docker run \
  --rm \
  -v "$OUTPUT":/host \
  -e "SANITY_CHECK=true" \
  sourcegraph/alpine@sha256:ce099fbcd3cf70b338fc4cb2a4e1fa9ae847de21afdb0a849a393b87d94fb174 \
  /host/symbols
