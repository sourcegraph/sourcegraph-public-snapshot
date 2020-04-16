#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/jaeger-collector:latest@sha256:81dc49f8d8abcf0dd116e096027a19b54d3c0d441164a71b44585ee0ee1095dc
docker tag index.docker.io/sourcegraph/jaeger-collector:latest@sha256:81dc49f8d8abcf0dd116e096027a19b54d3c0d441164a71b44585ee0ee1095dc $IMAGE
