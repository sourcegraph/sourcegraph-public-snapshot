#!/usr/bin/env bash

set -ex

if [ -z "$OSS_WEBAPP_VERSION" ]; then
    export OSS_WEBAPP_VERSION=$(npm info @sourcegraph/webapp version)
fi

# Generate and upload the pipeline.
export GO111MODULE=on
go mod edit -dropreplace github.com/sourcegraph/sourcegraph
go run ./dev/ci/gen-pipeline.go | buildkite-agent pipeline upload
