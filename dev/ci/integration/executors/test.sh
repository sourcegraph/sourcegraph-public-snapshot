#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

set -e

export SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"
export SRC_LOG_LEVEL=dbug

aspectRC="/tmp/aspect-generated.bazelrc"
rosetta bazelrc > "$aspectRC"

echo "--- run tests"
bazel --bazelrc="$aspectRC" run //dev/ci/integration/executors/tester:tester
