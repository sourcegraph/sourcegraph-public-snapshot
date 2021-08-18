#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-cache:21-08-13_c096e0d35@sha256:2e41070dbe579686ce421e556dd648f8d5a489505b9842262ba5a8b8a35fbd6c
docker tag index.docker.io/sourcegraph/redis-cache:21-08-13_c096e0d35a@sha256:2e41070dbe579686ce421e556dd648f8d5a489505b9842262ba5a8b8a35fbd6c "$IMAGE"
