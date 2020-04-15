#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/jaeger-agent:1.17.1@sha256:03ddce35d7bdbacdedc90369df84395e9aecfa51571ef661e5128ac1e95b8d5c
docker tag index.docker.io/sourcegraph/jaeger-agent:1.17.1@sha256:03ddce35d7bdbacdedc90369df84395e9aecfa51571ef661e5128ac1e95b8d5c $IMAGE
