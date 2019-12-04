#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

OUTPUT=`mktemp -d -t sgdockerbuild_XXXXXXX`
cleanup() {
    rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

for pkg in github.com/sourcegraph/sourcegraph/cmd/github-proxy; do
    go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION" -buildmode exe -tags dist -o $OUTPUT/$(basename $pkg) $pkg
done

docker build -f cmd/github-proxy/Dockerfile -t $IMAGE $OUTPUT \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
