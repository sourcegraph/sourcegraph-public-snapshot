#!/bin/bash

# This script is called by a step in the sourcegraph/deploy-sourcegraph-cloud pipeline, to run the codeintel-qa test suite
# against the preprod.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)

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

echo '--- :go: Building init-sg'
go build -o init-sg ./internal/cmd/init-sg/...

echo '--- :horse: Running init-sg addRepos'
./init-sg addRepos -config ./dev/ci/integration/code-intel/repos.json

echo '--- Installing local src-cli'
./dev/ci/integration/code-intel/install-src.sh
which src
src version

echo '--- :brain: Running the test suite'
pushd dev/codeintel-qa

echo '--- :zero: downloading test data from GCS'
go run ./cmd/download

echo '--- :one: clearing existing state'
go run ./cmd/clear

echo '--- :two: integration test ./dev/codeintel-qa/cmd/upload'
env PATH="${root_dir}/.bin:${PATH}" go run ./cmd/upload --timeout=5m

# make queries but do not assert against expected locations
echo '--- :three: integration test ./dev/codeintel-qa/cmd/query'
go run ./cmd/query -check-query-result=false
popd
