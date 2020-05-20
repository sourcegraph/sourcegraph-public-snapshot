#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:ff37f90@sha256:aa93514b7bc3aaf7a4e9c92e5ff52ee5052db6fb101255a69f054e5b8cdb46ff
docker tag index.docker.io/sourcegraph/syntect_server:ff37f90@sha256:aa93514b7bc3aaf7a4e9c92e5ff52ee5052db6fb101255a69f054e5b8cdb46ff "$IMAGE"
