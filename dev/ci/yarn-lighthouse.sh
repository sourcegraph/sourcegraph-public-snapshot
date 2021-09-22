#!/usr/bin/env bash

set -e

BASE_URL=http://localhost:3443
TEST_PATH=$1

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'lighthouse:prep'
tar -xf client.tar.gz -C .

echo "--- Yarn install in root"
# mutex is necessary since CI runs various yarn installs in parallel
NODE_ENV='' yarn --mutex network --frozen-lockfile --network-timeout 60000

echo "--- Collecting Lighthouse results"
yarn lighthouse collect --url="$BASE_URL$TEST_PATH"

echo "--- Uploading Lighthouse results"
yarn lighthouse upload --target=temporary-public-storage
