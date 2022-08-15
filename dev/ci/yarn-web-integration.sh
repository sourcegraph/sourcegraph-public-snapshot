#!/usr/bin/env bash

set -e

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'puppeteer:prep'
tar -xf client.tar.gz -C .

echo "--- Yarn install in root"
./dev/ci/yarn-install-with-retry.sh

echo "--- Run integration test suite"
yarn percy exec --quiet -- yarn _cover-integration "$@"

echo "--- Process NYC report"
yarn nyc report -r json

echo "--- Upload coverage report"
dev/ci/codecov.sh -c -F typescript -F integration
