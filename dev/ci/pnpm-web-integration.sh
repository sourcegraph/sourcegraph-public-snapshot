#!/usr/bin/env bash

set -e

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'puppeteer:prep'
tar -xf client.tar.gz -C .

echo "--- Pnpm install in root"
./dev/ci/pnpm-install-with-retry.sh

echo "--- Run integration test suite"
pnpm pnpm _test-integration "$@"
