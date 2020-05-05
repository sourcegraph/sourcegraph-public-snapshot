#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-cache:20-02-03_da9d71ca@sha256:7820219195ab3e8fdae5875cd690fed1b2a01fd1063bd94210c0e9d529c38e56
docker tag index.docker.io/sourcegraph/redis-cache:20-02-03_da9d71ca@sha256:7820219195ab3e8fdae5875cd690fed1b2a01fd1063bd94210c0e9d529c38e56 "$IMAGE"
