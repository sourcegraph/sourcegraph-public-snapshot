#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

set -e

export SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"
export SRC_LOG_LEVEL=dbug

pushd dev/ci/integration/executors 1>/dev/null
go run ./tester
popd 1>/dev/null
