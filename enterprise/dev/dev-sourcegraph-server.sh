#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/..

# Build a Sourcegraph server docker image with private code built in to run for
# development purposes
time cmd/server/pre-build.sh
IMAGE=sourcegraph/server:0.0.0-ENTERPRISE-DEVELOPMENT VERSION=0.0.0-ENTERPRISE-DEVELOPMENT time cmd/server/build.sh

IMAGE=sourcegraph/server:0.0.0-ENTERPRISE-DEVELOPMENT ../dev/run-server-image.sh
