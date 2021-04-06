#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/postgres-12.6:21-03-26_5d7084279@sha256:aa937c1c8ab20f3c809f04480d5a73791b05be59d3183726fd499ae0a123e982
docker tag index.docker.io/sourcegraph/postgres-12.6:21-03-26_5d7084279@sha256:aa937c1c8ab20f3c809f04480d5a73791b05be59d3183726fd499ae0a123e982 "$IMAGE"
