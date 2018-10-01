#!/usr/bin/env bash

set -ex

# Generate and upload the pipeline.
export GO111MODULE=on
go mod edit -dropreplace github.com/sourcegraph/sourcegraph
go run ./dev/ci/gen-pipeline.go | buildkite-agent pipeline upload
