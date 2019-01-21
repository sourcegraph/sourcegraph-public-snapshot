#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

# Additional images passed in here when this script is called externally by our
# enterprise build scripts.
ADDITIONAL_IMAGES=${@:-github.com/sourcegraph/sourcegraph/cmd/frontend}

# Overridable server package path for when this script is called externally by
# our enterprise build scripts.
SERVER_PKG=${SERVER_PKG:-github.com/sourcegraph/sourcegraph/cmd/server}

docker build \
       --build-arg VERSION="$VERSION" \
       --build-arg ADDITIONAL_IMAGES="$ADDITIONAL_IMAGES" \
       --build-arg SERVER_PKG="$SERVER_PKG" \
       -f cmd/server/Dockerfile \
       -t $IMAGE \
       .
