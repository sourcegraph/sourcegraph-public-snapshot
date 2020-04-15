#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/jaeger-query:latest@sha256:21a30fbaeed1290bd1fe9c93f67bd6c1d47b8d854ebc64808a6611a501991a58
docker tag index.docker.io/sourcegraph/jaeger-query:latest@sha256:21a30fbaeed1290bd1fe9c93f67bd6c1d47b8d854ebc64808a6611a501991a58 $IMAGE
