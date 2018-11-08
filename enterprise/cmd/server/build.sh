#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

GO111MODULE=on GOBIN=$PWD/.bin go install github.com/sourcegraph/godockerize

# Additional images passed in here when this script is called externally by our
# enterprise build scripts.
additional_images=${@:-github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend}

# Overridable server package path for when this script is called externally by
# our enterprise build scripts.
server_pkg=${SERVER_PKG:-github.com/sourcegraph/sourcegraph/enterprise/cmd/server}

GO111MODULE=on ./.bin/godockerize build --base 'alpine:3.8' -t ${IMAGE} --go-build-flags="-ldflags" --go-build-flags="-X github.com/sourcegraph/sourcegraph/pkg/version.version=${VERSION}" --env VERSION=${VERSION} \
    $server_pkg \
    github.com/sourcegraph/sourcegraph/cmd/management-console \
    github.com/sourcegraph/sourcegraph/cmd/github-proxy \
    github.com/sourcegraph/sourcegraph/cmd/gitserver \
    github.com/sourcegraph/sourcegraph/cmd/query-runner \
    github.com/sourcegraph/sourcegraph/cmd/symbols \
    github.com/sourcegraph/sourcegraph/cmd/repo-updater \
    github.com/sourcegraph/sourcegraph/cmd/searcher \
    github.com/sourcegraph/sourcegraph/cmd/indexer \
    github.com/google/zoekt/cmd/zoekt-archive-index \
    github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
    github.com/google/zoekt/cmd/zoekt-webserver \
    github.com/sourcegraph/sourcegraph/cmd/lsp-proxy $additional_images
