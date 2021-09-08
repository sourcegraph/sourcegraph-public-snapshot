#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:21-08-31_c330964@sha256:759f331a474d2a67b811a1b374b0b24a4661446a2d8e9b211f51ea8ae95e1130
docker tag docker.io/sourcegraph/syntect_server:21-08-31_c330964@sha256:759f331a474d2a67b811a1b374b0b24a4661446a2d8e9b211f51ea8ae95e1130 "$IMAGE"
