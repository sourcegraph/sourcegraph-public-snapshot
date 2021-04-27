#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:c44c428@sha256:ae234db20fedefcab06d89079f71ecd0e6b1f91f3d8aee96ef0935e2a79599df
docker tag index.docker.io/sourcegraph/syntect_server:c44c428@sha256:ae234db20fedefcab06d89079f71ecd0e6b1f91f3d8aee96ef0935e2a79599df "$IMAGE"
