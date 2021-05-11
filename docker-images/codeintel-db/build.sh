#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/codeintel-db/build.sh
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/postgres-12.6:95244_2021-05-06_2c1f77e@sha256:35040317490324a15e1259c9023a726eb27c694530f6f8877e87d337c7b97778
docker tag index.docker.io/sourcegraph/postgres-12.6:95244_2021-05-06_2c1f77e@sha256:35040317490324a15e1259c9023a726eb27c694530f6f8877e87d337c7b97778 "$IMAGE"
