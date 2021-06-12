#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-store:21-06-10_6ded3b35f@sha256:5e9fec6f4f9801fcd012a5eb4d35d1f50e9e7c6aad2733790698ab28b8378930
docker tag index.docker.io/sourcegraph/redis-store:21-06-10_6ded3b35f@sha256:5e9fec6f4f9801fcd012a5eb4d35d1f50e9e7c6aad2733790698ab28b8378930 "$IMAGE"
