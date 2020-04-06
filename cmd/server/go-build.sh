#!/usr/bin/env bash

set -euxo pipefail

PACKAGE="$1"

# Some packages require additional build arguments (e.g. CGO).
# Sourcing this file will add additional values to the environment.
if [[ -f "cmd/$(basename "$PACKAGE")/go-build-args.sh" ]]; then
    . "cmd/$(basename "$PACKAGE")/go-build-args.sh"
fi

go build \
    -trimpath \
    -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION"  \
    -buildmode exe \
    -installsuffix netgo \
    -tags "dist netgo" \
    -o "$BINDIR/$(basename "$PACKAGE")" "$PACKAGE"
