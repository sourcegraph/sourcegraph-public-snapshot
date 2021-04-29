#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:453e9ca@sha256:c3f19209794119c703f43a830918005cd6346e15d6db56e42dde63f7acb23b9e
docker tag index.docker.io/sourcegraph/syntect_server:453e9ca@sha256:c3f19209794119c703f43a830918005cd6346e15d6db56e42dde63f7acb23b9e "$IMAGE"
