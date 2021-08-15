#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-store:21-08-13_c096e0d35@sha256:3afb95c520eb9f32cc02b4f1d1208ff68360411a1b992a056aeb6020eb95499e
docker tag index.docker.io/sourcegraph/redis-store:21-08-13_c096e0d35@sha256:3afb95c520eb9f32cc02b4f1d1208ff68360411a1b992a056aeb6020eb95499e "$IMAGE"
