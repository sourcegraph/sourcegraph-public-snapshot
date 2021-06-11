#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-cache:21-06-10_6ded3b35fa@sha256:35245d84f0d154d12501ea4b997fa21913939cc4ce7c4de288b44d1d7e50624c
docker tag index.docker.io/sourcegraph/redis-cache:21-06-10_6ded3b35fa@sha256:35245d84f0d154d12501ea4b997fa21913939cc4ce7c4de288b44d1d7e50624c "$IMAGE"
