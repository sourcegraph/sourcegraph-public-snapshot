#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:33dc1ba@sha256:625c556f5cf456144e51e3fe55e6312398b7714994165b4f605711e6f7d862a0
docker tag index.docker.io/sourcegraph/syntect_server:33dc1ba@sha256:625c556f5cf456144e51e3fe55e6312398b7714994165b4f605711e6f7d862a0 "$IMAGE"
