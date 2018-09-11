#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

rm -f ./dockerfile.go
cp ../cmd/server/dockerfile.go .

# TODO: These should all be under the enterprise folder/repo
SERVER_PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/server ../cmd/server/build.sh \
    github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend \
    github.com/sourcegraph/sourcegraph/cmd/indexer \
    github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-archive-index \
    github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
    github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-webserver
