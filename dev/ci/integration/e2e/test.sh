#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
set -e

URL="${1:-"http://localhost:7080"}"

echo "--- yarn run test-e2e"
env SOURCEGRAPH_BASE_URL="$URL" yarn run cover-e2e

echo "--- coverage"
yarn nyc report -r json
# Upload the coverage under the "e2e" flag (toggleable in the CodeCov UI)
./dev/ci/codecov.sh -F e2e
