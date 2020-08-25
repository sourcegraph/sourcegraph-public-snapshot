#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:c22bde0@sha256:07b9f1ff4bd2c60299f9404144cd72897fa4de2308d1be65c35bcdcd10e5410d
docker tag index.docker.io/sourcegraph/syntect_server:c22bde0@sha256:07b9f1ff4bd2c60299f9404144cd72897fa4de2308d1be65c35bcdcd10e5410d "$IMAGE"
