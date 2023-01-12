#!/usr/bin/env bash

# This script builds the frontend docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

# Environment for building linux binaries
#export GOARCH=amd64
#export GOOS=linux

# macOS
#export GOARCH=arm64
#export GOOS=darwin
#export CGO_ENABLED=1
#export CC=clang
#export CGO_CFLAGS='-target arm64-apple-darwin21.6.0'

export GOOS=$(go env GOOS)
export GOARCH=$(go env GOARCH)

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph"

ENTERPRISE=1 DEV_WEB_BUILDER=esbuild yarn run build-web
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=${VERSION-0.0.0+dev} -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s) -X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=single-program" -buildmode exe -tags dist -o ".bin/$(basename $pkg)-$GOOS-$GOARCH-dist" "$pkg"
