#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:9c1c9d7@sha256:c7a5d90ad995181ad520535c9b4d809576514933bbe60b625f821d78c31628cd
docker tag index.docker.io/sourcegraph/syntect_server:9c1c9d7@sha256:c7a5d90ad995181ad520535c9b4d809576514933bbe60b625f821d78c31628cd "$IMAGE"
