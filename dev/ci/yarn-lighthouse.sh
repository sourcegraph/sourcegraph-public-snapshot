#!/usr/bin/env bash

set -e

BASE_URL=http://localhost:3443
TEST_LABEL=$1
TEST_PATH=$2
[[ "$BUILDKITE_BRANCH" == "tr/lighthouse-ci" ]] && SHOULD_STORE_RESULTS='true' || SHOULD_STORE_RESULTS='false'

echo Storing results: $SHOULD_STORE_RESULTS

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'lighthouse:prep'
tar -xf client.tar.gz -C .

echo "--- Yarn install in root"
# mutex is necessary since CI runs various yarn installs in parallel
NODE_ENV='' yarn --mutex network --frozen-lockfile --network-timeout 60000

echo "--- Collecting Lighthouse results"
yarn lhci collect --url="$BASE_URL$TEST_PATH"

echo "--- Uploading Lighthouse results"
yarn lhci upload --githubStatusContextSuffix="/$TEST_LABEL" --uploadUrlMap="$SHOULD_STORE_RESULTS"
