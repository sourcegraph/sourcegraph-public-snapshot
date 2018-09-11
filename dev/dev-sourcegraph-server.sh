#!/bin/sh

set -ex

# Build a Sourcegraph server docker image to run for development purposes. Note
# that this image is not exactly identical to the published sourcegraph/server
# images, as those include Sourcegraph's proprietary code behind paywalls.
time cmd/server/pre-build.sh
IMAGE=sourcegraph/server:$USER-dev VERSION=$USER-dev time cmd/server/build.sh

IMAGE=sourcegraph/server:$USER-dev ${BASH_SOURCE%/*}/run-server-image.sh
