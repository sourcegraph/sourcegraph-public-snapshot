#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:9089f98@sha256:83ff65809e6647b466bd400de4c438a32feeabe8e791b12e15c67c84529ad2de
docker tag index.docker.io/sourcegraph/syntect_server:9089f98@sha256:83ff65809e6647b466bd400de4c438a32feeabe8e791b12e15c67c84529ad2de "$IMAGE"
