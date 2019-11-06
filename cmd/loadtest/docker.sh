#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -euxo pipefail

docker build -f cmd/loadtest/Dockerfile -t $IMAGE . \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
