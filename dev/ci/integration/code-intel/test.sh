#!/usr/bin/env bash

# This script runs the codeintel-qa tests against a running server.
# This script is invoked by ./dev/ci/integration/run-integration.sh after running an instance.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)

SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"
export SOURCEGRAPH_BASE_URL

echo '--- :go: Building init-sg'
go build -o init-sg ./internal/cmd/init-sg/...

echo '--- Initializing instance'
"${root_dir}/init-sg" initSG

echo '--- Loading secrets'
set +x # Avoid printing secrets
# shellcheck disable=SC1091
source /root/.sg_envrc
set -x

echo '--- :horse: Running init-sg addRepos'
"${root_dir}/init-sg" addRepos -config ./dev/ci/integration/code-intel/repos.json

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

echo '--- :three: integration test ./dev/codeintel-qa/cmd/query'
go run ./cmd/query
popd
