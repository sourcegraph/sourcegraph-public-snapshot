#!/usr/bin/env bash

set -euxo pipefail

PACKAGE="$1"

if [[ -f "cmd/$(basename "$PACKAGE")/go-build.sh" ]]; then
    # Application builds itself (e.g. requires CGO)
    bash "cmd/$(basename "$PACKAGE")/go-build.sh"
else
    go build \
        -trimpath \
        -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION"  \
        -buildmode exe \
        -installsuffix netgo \
        -tags "dist netgo" \
        -o "$BINDIR/$(basename "$PACKAGE")" "$PACKAGE"
fi
