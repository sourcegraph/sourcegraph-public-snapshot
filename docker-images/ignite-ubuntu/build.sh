#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Build the ignite-ubuntu image for use in firecracker. Set SRC_CLI_VERSION to the minimum required version in internal/src-cli/consts.go.
docker build -t "${IMAGE:-sourcegraph/ignite-ubuntu}" --build-arg SRC_CLI_VERSION="$(go run ../../internal/cmd/src-cli-version/main.go)" .
