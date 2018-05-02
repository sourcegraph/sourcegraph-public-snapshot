#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

GOBIN=$PWD/vendor/.bin go install ./vendor/github.com/sourcegraph/godockerize

./vendor/.bin/godockerize build --base alpine:3.6 -t ${IMAGE} --env VERSION=${VERSION} \
			  github.com/sourcegraph/sourcegraph/cmd/server \
			  github.com/sourcegraph/sourcegraph/cmd/frontend \
			  github.com/sourcegraph/sourcegraph/cmd/github-proxy \
			  github.com/sourcegraph/sourcegraph/cmd/gitserver \
			  github.com/sourcegraph/sourcegraph/cmd/indexer \
			  github.com/sourcegraph/sourcegraph/cmd/query-runner \
			  github.com/sourcegraph/sourcegraph/cmd/symbols \
			  github.com/sourcegraph/sourcegraph/cmd/repo-updater \
			  github.com/sourcegraph/sourcegraph/cmd/searcher \
			  github.com/sourcegraph/sourcegraph/cmd/lsp-proxy \
			  github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-archive-index \
			  github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
			  github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-webserver
