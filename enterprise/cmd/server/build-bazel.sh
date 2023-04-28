#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../../..

ENTERPRISE=true ./cmd/server/build-bazel.sh
