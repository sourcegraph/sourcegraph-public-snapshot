#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/..
set -ex

# Build a Sourcegraph server docker image with private code built in to run for
# development purposes
time ../sourcegraph/cmd/server/pre-build.sh
IMAGE=sourcegraph/server:$USER-dev VERSION=$USER-dev time cmd/server/build.sh

IMAGE=sourcegraph/server:$USER-dev ../sourcegraph/dev/run-server-image.sh
