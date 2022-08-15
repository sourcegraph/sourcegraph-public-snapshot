#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../../../../..
set -eu

SRC_CLI_VERSION="$(go run ./internal/cmd/src-cli-version/main.go)"

docker build --platform linux/amd64 -t "sourcegraph/ignite-ubuntu:insiders" --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" ./enterprise/cmd/executor/vm-image/ignite-ubuntu
