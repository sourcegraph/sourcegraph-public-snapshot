#!/usr/bin/env bash

set -ex

if [ -z "$OSS_WEBAPP_VERSION" ]; then
    export OSS_WEBAPP_VERSION=$(npm info @sourcegraph/webapp version)
fi

# Generate and upload the pipeline.
go run ./dev/ci/gen-pipeline.go | buildkite-agent pipeline upload
