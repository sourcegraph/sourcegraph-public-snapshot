#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
set -e

# URL="${1:-"http://localhost:7080"}"

echo "--- bazel test e2e"
bazel \
  --bazelrc=.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  test \
  //client/web/src/end-to-end:e2e
