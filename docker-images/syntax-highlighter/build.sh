#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:f68be78@sha256:1d2ac738eec37f8a3ac4da3d73350a4f9be6a2d730074f07ce42dd9dd978b5fc
docker tag docker.io/sourcegraph/syntect_server:f68be78@sha256:1d2ac738eec37f8a3ac4da3d73350a4f9be6a2d730074f07ce42dd9dd978b5fc "$IMAGE"
