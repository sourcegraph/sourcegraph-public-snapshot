#!/usr/bin/env bash

# This script runs the codeintel-qa tests against a running server.
# This script is invoked by ./dev/ci/integration/run-integration.sh after running an instance.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"
export SOURCEGRAPH_BASE_URL

echo "--- Loading secrets"
set +x # Avoid printing secrets
# shellcheck disable=SC1091
source /root/.sg_envrc
set -x

echo "--- :go: Building init-sg"
go build -o init-sg ./internal/cmd/init-sg/...

# TODO - download src-cli

echo "--- :horse: Running init-sg addRepos"
./init-sg addRepos -config ./dev/ci/integration/code-intel/repos.json

echo "--- :brain: Running the test suite"
pushd dev/codeintel-qa

echo '--- :zero: downloading test data from GCS'
go run ./cmd/download

echo '--- :one: clearing existing state'
go run ./cmd/clear

# Disable migration #20 (LSIF -> SCIP)
echo '--- :two: Disabling LSIF -> SCIP migration'
./init-sg oobmigration -id T3V0T2ZCYW5kTWlncmF0aW9uOjIw -down

echo '--- :three: integration test ./dev/codeintel-qa/cmd/upload'
go run ./cmd/upload --timeout=5m

echo '--- :four: integration test ./dev/codeintel-qa/cmd/query'
go run ./cmd/query

# Enable migration #20 (LSIF -> SCIP) and wait for it to complete
echo '--- :five: Running LSIF -> SCIP migration'
./init-sg oobmigration -id T3V0T2ZCYW5kTWlncmF0aW9uOjIw

echo '--- :six: integration test ./dev/codeintel-qa/cmd/query'
go run ./cmd/query
popd
