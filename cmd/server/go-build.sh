#!/usr/bin/env bash

set -euxo pipefail

PACKAGE="$1"

basepkg="$(basename "$PACKAGE")"
if [[ "$basepkg" != "server" ]] && [[ -f "cmd/$basepkg/go-build.sh" ]]; then
  # Application builds itself (e.g. requires CGO)
  bash "cmd/$basepkg/go-build.sh" "$BINDIR"
else
  go build \
    -trimpath \
    -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION" \
    -buildmode exe \
    -installsuffix netgo \
    -tags "dist netgo" \
    -o "$BINDIR/$basepkg" "$PACKAGE"
fi
