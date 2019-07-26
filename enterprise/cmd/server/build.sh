#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -ex

export SERVER_PKG=${SERVER_PKG:-sourcegraph.com/enterprise/cmd/server}

./cmd/server/build.sh sourcegraph.com/enterprise/cmd/frontend sourcegraph.com/enterprise/cmd/management-console
