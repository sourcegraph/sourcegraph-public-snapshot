#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:21-08-31_6556204@sha256:e1a17d5b9abd7b4b1c3cd6a0c12e2b686598dd679760efb85fdc4148faf4eb75
docker tag docker.io/sourcegraph/syntect_server:21-08-31_6556204@sha256:e1a17d5b9abd7b4b1c3cd6a0c12e2b686598dd679760efb85fdc4148faf4eb75 "$IMAGE"
