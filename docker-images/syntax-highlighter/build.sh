#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:21-06-11_edf94dc@sha256:d3414326b4e1f3875c6f5558de9160dc4140615061b2ef45c99ac7059d167dec
docker tag index.docker.io/sourcegraph/syntect_server:21-06-11_edf94dc@sha256:d3414326b4e1f3875c6f5558de9160dc4140615061b2ef45c99ac7059d167dec "$IMAGE"
