#!/usr/bin/env bash

set -euxo pipefail

PACKAGE="$1"

go build \
    -trimpath \
    -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION"  \
    -buildmode exe \
    -installsuffix netgo \
    -tags "dist netgo" \
    -o "$BINDIR/$(basename "$PACKAGE")" "$PACKAGE"
