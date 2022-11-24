#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -e

export SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"

pushd dev/ci/integration/executors
go run ./cmd/...
popd || exit 1
