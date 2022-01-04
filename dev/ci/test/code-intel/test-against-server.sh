#!/usr/bin/env bash

# This script runs the codeintel-qa tests against a running server.
# This script is invoked by ./dev/ci/run-integration.sh after running an instance.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
SG_ROOT=$(pwd)
set -ex

export SOURCEGRAPH_BASE_URL="$1"

echo '--- initializing Sourcegraph instance'

pushd internal/cmd/init-sg || exit 1
go build -o "${SG_ROOT}/init-sg"
popd || exit 1

pushd dev/ci/test/code-intel || exit 1
"${SG_ROOT}/init-sg" initSG
# Disable `-x` to avoid printing secrets
set +x
# shellcheck disable=SC1091
source /root/.sg_envrc
set -x
"${SG_ROOT}/init-sg" addRepos -config repos.json
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
