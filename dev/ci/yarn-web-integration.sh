#!/usr/bin/env bash

set -e

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'puppeteer:prep'
tar -xf client.tar.gz -C .

echo "--- Yarn install in root"
# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --network-timeout 60000

echo "--- Download Puppeteer browser"
yarn --cwd client/shared run download-puppeteer-browser

echo "--- Run integration test suite"
# Word splittinng is intentional here. $1 contains a string with test files separated by a space.
# shellcheck disable=SC2086
yarn percy exec --parallel yarn cover-integration:base $1

echo "--- Process NYC report"
yarn nyc report -r json

echo "--- Upload coverage report"
dev/ci/codecov.sh -c -F typescript -F integration
