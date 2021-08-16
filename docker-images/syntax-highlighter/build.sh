#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:21-08-13_a2d6b30@sha256:b6f9bd9c1da621b0109855fd9268bf54c8537b5b4521d720d71e3cb244a2e336
docker tag docker.io/sourcegraph/syntect_server:21-08-13_a2d6b30@sha256:b6f9bd9c1da621b0109855fd9268bf54c8537b5b4521d720d71e3cb244a2e336 "$IMAGE"
