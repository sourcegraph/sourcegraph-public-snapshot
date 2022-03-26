#!/usr/bin/env bash

set -e

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'puppeteer:prep'
tar -xf client.tar.gz -C .

echo "--- pnpm install in root"
# mutex is necessary since CI runs various pnpm installs in parallel
pnpm install

echo "--- Run integration test suite"
pnpm percy exec --parallel pnpm cover-integration:base "$@"

echo "--- Process NYC report"
pnpm nyc report -r json

echo "--- Upload coverage report"
dev/ci/codecov.sh -c -F typescript -F integration
