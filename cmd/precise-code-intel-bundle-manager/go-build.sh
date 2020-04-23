#!/usr/bin/env bash

# This script builds the precise-code-intel-bundle-manager go binary.
# Requires a single argument which is the path to the target bindir.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT="${1:?no output path provided}"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux

# Get additional build args
. ./dev/libsqlite3-pcre/go-build-args.sh

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"
