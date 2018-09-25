#!/usr/bin/env bash

set -ex

# Clone the open source repository first (gen-pipeline depends on it).
./dev/ci/ensure-oss-repo-cloned.sh

# Make sure all steps use the same OSS revision and version of @sourcegraph/webapp,
# no matter if on master or on a branch
if [ -z "$OSS_REPO_REVISION" ]; then
    pushd $GOPATH/src/github.com/sourcegraph/sourcegraph
    export OSS_REPO_REVISION=$(git rev-parse HEAD)
    popd
fi
if [ -z "$OSS_WEBAPP_VERSION" ]; then
    export OSS_WEBAPP_VERSION=$(npm info @sourcegraph/webapp version)
fi

# Generate and upload the pipeline.
go run ./dev/ci/gen-pipeline.go | buildkite-agent pipeline upload
