#!/usr/bin/env bash

set -euxo pipefail

PACKAGE="$1"
RELATIVE_PACKAGE="${PACKAGE#github.com/sourcegraph/sourcegraph/}"
BASENAME="$(basename "$PACKAGE")"

if [[ "$BASENAME" != "server" ]] && [[ -f "$RELATIVE_PACKAGE/go-build.sh" ]]; then
  # Application builds itself (e.g. requires CGO)
  bash "$RELATIVE_PACKAGE/go-build.sh" "$BINDIR"
else
  go build \
    -trimpath \
    -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
    -buildmode exe \
    -installsuffix netgo \
    -tags "dist netgo" \
    -o "$BINDIR/$BASENAME" "$PACKAGE"
fi
