#!/usr/bin/env bash

set -e

BASE_URL=http://localhost:3443

# Store results on `main` so future branches can be compared against a baseline
[[ "$BUILDKITE_BRANCH" == "main" ]] && SHOULD_STORE_RESULTS='true' || SHOULD_STORE_RESULTS='false'

echo "--- Download pre-built client artifact"
buildkite-agent artifact download 'client.tar.gz' . --step 'lighthouse:prep'
tar -xf client.tar.gz -C .

echo "--- Yarn install in root"
# mutex is necessary since CI runs various yarn installs in parallel
NODE_ENV='' yarn --mutex network --frozen-lockfile --network-timeout 60000

echo "--- Runing Lighthouse"
yarn lhci autorun --url="$BASE_URL/" --url="$BASE_URL/search?q=repo:sourcegraph/lighthouse-ci-test-repository+file:index.js" --url="$BASE_URL/github.com/sourcegraph/lighthouse-ci-test-repository" --url="$BASE_URL/github.com/sourcegraph/lighthouse-ci-test-repository/-/blob/index.js"
