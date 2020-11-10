#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/syntect_server:1c72add@sha256:17851d7da26f86e9f21725ac461b3ce6066bdb95731da09bed07045c03661fe9
docker tag index.docker.io/sourcegraph/syntect_server:1c72add@sha256:17851d7da26f86e9f21725ac461b3ce6066bdb95731da09bed07045c03661fe9 "$IMAGE"
