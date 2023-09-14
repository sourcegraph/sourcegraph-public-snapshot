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
# env SOURCEGRAPH_BASE_URL="$URL" pnpm run cover-e2e

# echo "--- coverage"
# pnpm nyc report -r json
# # Upload the coverage under the "e2e" flag (toggleable in the CodeCov UI)
# ./dev/ci/codecov.sh -F e2e
