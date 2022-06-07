#!/usr/bin/env bash

# This script runs the codeintel-qa tests against a running server.
# This script is invoked by ./dev/ci/integration/run-integration.sh after running an instance.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -e

export SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"

echo '--- initializing Sourcegraph instance'

pushd internal/cmd/init-sg || exit 1
go build -o "${root_dir}/init-sg"
popd || exit 1

pushd dev/ci/integration/code-intel || exit 1
"${root_dir}/init-sg" initSG
# Disable `-x` to avoid printing secrets
set +x
# shellcheck disable=SC1091
source /root/.sg_envrc
"${root_dir}/init-sg" addRepos -config repos.json
popd || exit 1

pushd dev/codeintel-qa || exit 1
echo '--- downloading test data from GCS'
./scripts/download.sh
echo '--- integration test ./dev/codeintel-qa/cmd/upload'
go build ./cmd/upload
./upload --timeout=5m
echo '--- integration test ./dev/codeintel-qa/cmd/query'
go build ./cmd/query
./query
popd || exit 1
