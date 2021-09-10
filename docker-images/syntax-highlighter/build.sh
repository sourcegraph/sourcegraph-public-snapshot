#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:21-09-10_c4f947f@sha256:f15a5dcc88ab8574049e37c9985750d0a4aa3d1ec665ec8345f85206155364fb
docker tag docker.io/sourcegraph/syntect_server:21-09-10_c4f947f@sha256:f15a5dcc88ab8574049e37c9985750d0a4aa3d1ec665ec8345f85206155364fb "$IMAGE"
