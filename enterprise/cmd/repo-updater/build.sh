#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../../..

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then
  ./cmd/repo-updater/build.sh //enterprise/cmd/repo-updater
  exit $?
fi

./cmd/repo-updater/build.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater
