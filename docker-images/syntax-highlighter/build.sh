#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:331beda@sha256:6b8950b41993af3d10300b5a160fab1c06c337cb4614978f4fdb76e2588afbfe
docker tag index.docker.io/sourcegraph/syntect_server:331beda@sha256:6b8950b41993af3d10300b5a160fab1c06c337cb4614978f4fdb76e2588afbfe "$IMAGE"
