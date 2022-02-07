#!/usr/bin/env bash

set -e

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'puppeteer:prep'
tar -xf client.tar.gz -C .

echo "--- Yarn install in root"
# mutex is necessary since CI runs various yarn installs in parallel
pnpm --frozen-lockfile

echo "--- Run integration test suite"
# Word splittinng is intentional here. $1 contains a string with test files separated by a space.
# shellcheck disable=SC2086
pnpm percy exec --parallel pnpm cover-integration:base $1

echo "--- Process NYC report"
pnpm nyc report -r json

echo "--- Upload coverage report"
dev/ci/codecov.sh -c -F typescript -F integration
