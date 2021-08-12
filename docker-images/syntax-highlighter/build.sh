#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:32d880d@sha256:899661691c3a6f8d587186bed73c3224b065d1e1c3485aff2ea208c261c010f6
docker tag docker.io/sourcegraph/syntect_server:32d880d@sha256:899661691c3a6f8d587186bed73c3224b065d1e1c3485aff2ea208c261c010f6 "$IMAGE"
