#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:3f75b2d@sha256:a5f9b8d8a78107310d17bd2041102a89324ff35ccf6769807084747912ea7eda
docker tag index.docker.io/sourcegraph/syntect_server:3f75b2d@sha256:a5f9b8d8a78107310d17bd2041102a89324ff35ccf6769807084747912ea7eda "$IMAGE"
