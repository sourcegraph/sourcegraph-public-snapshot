#!/bin/bash

# This script is called by a step in the sourcegraph/deploy-sourcegraph-cloud pipeline, to run the codeintel-qa test suite
# against the preprod.

set -eux

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

SOURCEGRAPH_BASE_URL="https://preview.sgdev.dev"
TEST_USER_EMAIL="testadmin@preview.sgdev.dev"
SOURCEGRAPH_SUDO_USER="admin"
TEST_USER_PASSWORD="$(gcloud secrets versions access latest --project=sourcegraph-ci --secret="PREPROD_TESTADMIN_PASSWORD" --quiet)"
SOURCEGRAPH_SUDO_TOKEN="$(gcloud secrets versions access latest --project=sourcegraph-ci --secret="PREPROD_TESTADMIN_SG_TOKEN" --quiet)"
export SOURCEGRAPH_BASE_URL
export TEST_USER_EMAIL
export SOURCEGRAPH_SUDO_USER
export TEST_USER_PASSWORD
export SOURCEGRAPH_SUDO_TOKEN

echo "--- :go: Building init-sg"
go build -o init-sg ./internal/cmd/init-sg/...

echo "--- :horse: Running init-sg addRepos"
./init-sg addRepos -config ./dev/ci/integration/code-intel/repos.json

pushd dev/codeintel-qa

echo "--- :brain: Running the test suite"
echo '--- :zero: downloading test data from GCS'
./scripts/download.sh
echo '--- :one: clearing existing state'
go run ./cmd/clear
echo '--- :two: integration test ./dev/codeintel-qa/cmd/upload'
go run ./cmd/upload --timeout=5m
echo '--- :three: integration test ./dev/codeintel-qa/cmd/query'
go run ./cmd/query -check-query-result=false # make queries but do not assert against expected locations
popd || exit 1
