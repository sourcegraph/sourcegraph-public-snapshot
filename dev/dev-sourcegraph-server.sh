#!/bin/sh

set -ex

# Build a Sourcegraph server docker image to run for development purposes
time cmd/server/pre-build.sh
time IMAGE=sourcegraph/server:$USER-dev VERSION=$USER-dev cmd/server/build.sh

sudo rm -rf /tmp/sourcegraph
docker run \
	-e SRC_LOG_LEVEL=dbug \
	--publish 7080:7080 \
	--rm \
	--volume /tmp/sourcegraph/config:/etc/sourcegraph \
	--volume /tmp/sourcegraph/data:/var/opt/sourcegraph \
	-v /var/run/docker.sock:/var/run/docker.sock \
	sourcegraph/server:$USER-dev
