#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:3342752@sha256:b9e1f7471ebe596415ca2c7ab8e1282d7c4ba4e4e71390d80e9924a73139d793
docker tag index.docker.io/sourcegraph/syntect_server:3342752@sha256:b9e1f7471ebe596415ca2c7ab8e1282d7c4ba4e4e71390d80e9924a73139d793 "$IMAGE"
