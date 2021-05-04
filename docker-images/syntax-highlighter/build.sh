#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:dd97058@sha256:d7163842f41388f41d19ce04833ac5f6d4e41d212869e7d2aea9c38ba6e77261
docker tag index.docker.io/sourcegraph/syntect_server:dd97058@sha256:d7163842f41388f41d19ce04833ac5f6d4e41d212869e7d2aea9c38ba6e77261 "$IMAGE"
