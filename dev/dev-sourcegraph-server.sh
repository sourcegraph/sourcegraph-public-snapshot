#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/..
set -ex

# Build a Sourcegraph server docker image to run for development purposes. Note
# that this image is not exactly identical to the published sourcegraph/server
# images, as those include Sourcegraph Enterprise features.
time cmd/server/pre-build.sh
IMAGE=sourcegraph/server:$USER-dev VERSION=$USER-dev time cmd/server/build.sh

IMAGE=sourcegraph/server:$USER-dev dev/run-server-image.sh
