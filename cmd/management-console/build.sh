#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -euxo pipefail

bindir="$OUTPUT_DIR/usr/local/bin"
mkdir -p "$bindir"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

echo "--- go build"
for pkg in $MANAGEMENT_CONSOLE_PKG; do
    go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION" -buildmode exe -tags dist -o "$bindir/$(basename "$pkg")" $pkg
done
