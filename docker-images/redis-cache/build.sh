#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/redis-cache:21-06-10_6ded3b35fa@sha256:e63e5dd4355bb9e042e032501cca6c54001cdb877dcd8c481fe57f9a542aa7be
docker tag pull index.docker.io/sourcegraph/redis-cache:21-06-10_6ded3b35fa@sha256:e63e5dd4355bb9e042e032501cca6c54001cdb877dcd8c481fe57f9a542aa7be "$IMAGE"
