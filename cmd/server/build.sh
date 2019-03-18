#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eux

OUTPUT=$(mktemp -d -t sgserver_XXXXXXX)
cleanup() {
    rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

# Additional images passed in here when this script is called externally by our
# enterprise build scripts.
additional_images=${@:-github.com/sourcegraph/sourcegraph/cmd/frontend github.com/sourcegraph/sourcegraph/cmd/management-console}

# Overridable server package path for when this script is called externally by
# our enterprise build scripts.
server_pkg=${SERVER_PKG:-github.com/sourcegraph/sourcegraph/cmd/server}

cp -a ./cmd/server/rootfs/. "$OUTPUT"
bindir="$OUTPUT/usr/local/bin"
mkdir -p "$bindir"

echo "--- go build"
for pkg in $server_pkg \
    github.com/sourcegraph/sourcegraph/cmd/github-proxy \
    github.com/sourcegraph/sourcegraph/cmd/gitserver \
    github.com/sourcegraph/sourcegraph/cmd/query-runner \
    github.com/sourcegraph/sourcegraph/cmd/repo-updater \
    github.com/sourcegraph/sourcegraph/cmd/searcher \
    github.com/google/zoekt/cmd/zoekt-archive-index \
    github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
    github.com/google/zoekt/cmd/zoekt-webserver $additional_images; do

    go build -ldflags "-X github.com/sourcegraph/sourcegraph/pkg/version.version=$VERSION" -buildmode exe -tags dist -o "$bindir/$(basename "$pkg")" "$pkg"
done

echo "--- build sqlite for symbols"
env CTAGS_D_OUTPUT_PATH="$OUTPUT/.ctags.d" SYMBOLS_EXECUTABLE_OUTPUT_PATH="$bindir/symbols" BUILD_TYPE=dist ./cmd/symbols/build.sh buildSymbolsDockerImageDependencies

echo "--- docker build"
docker build -f cmd/server/Dockerfile -t "$IMAGE" "$OUTPUT" --label "commitSHA=$COMMIT_SHA" --label "version=$VERSION"
