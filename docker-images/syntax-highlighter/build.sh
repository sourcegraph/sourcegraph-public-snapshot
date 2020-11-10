#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:b55ad2b@sha256:3e34d4996f046df081f5f9b726d992c73724ab1611ea55b104a937304faefdbe
docker tag index.docker.io/sourcegraph/syntect_server:b55ad2b@sha256:3e34d4996f046df081f5f9b726d992c73724ab1611ea55b104a937304faefdbe "$IMAGE"
