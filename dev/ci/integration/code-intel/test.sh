#!/usr/bin/env bash

# This script runs the codeintel-qa tests against a running server.
# This script is invoked by ./dev/ci/integration/run-integration.sh after running an instance.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)

SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"
export SOURCEGRAPH_BASE_URL

echo '--- :go: Building init-sg'
bazel build //internal/cmd/init-sg
out=$(bazel cquery //internal/cmd/init-sg --output=files)
cp "$out" "$root_dir/"

echo '--- Initializing instance'
"$root_dir/init-sg" initSG

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
bazel run //dev/codeintel-qa/cmd/download

echo '--- :one: clearing existing state'
bazel run //dev/codeintel-qa/cmd/clear

echo '--- :two: integration test ./dev/codeintel-qa/cmd/upload'
bazel run //dev/codeintel-qa/cmd/upload -- --timeout=5m --index-dir="$root_dir/dev/codeintel-qa/testdata/indexes"

echo '--- :three: integration test ./dev/codeintel-qa/cmd/query'
bazel run //dev/codeintel-qa/cmd/query -- --index-dir="$root_dir/dev/codeintel-qa/testdata/indexes"
popd
