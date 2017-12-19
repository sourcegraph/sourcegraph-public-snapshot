#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

GOBIN=$PWD/vendor/.bin go install ./vendor/github.com/sourcegraph/godockerize

./vendor/.bin/godockerize build -t ${IMAGE} --env VERSION=${VERSION} \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/server \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/github-proxy \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/indexer \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-list-updater \
			  sourcegraph.com/sourcegraph/sourcegraph/cmd/searcher
