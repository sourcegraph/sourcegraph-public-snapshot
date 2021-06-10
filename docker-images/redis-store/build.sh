#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-store:21-06-10_6ded3b35f@sha256:24005ea4a8fb906ff239e706bdb45e9c5838be8ec0287330b1138c1ec65562e8
docker tag index.docker.io/sourcegraph/redis-store:21-06-10_6ded3b35f@sha256:24005ea4a8fb906ff239e706bdb45e9c5838be8ec0287330b1138c1ec65562e8 "$IMAGE"
