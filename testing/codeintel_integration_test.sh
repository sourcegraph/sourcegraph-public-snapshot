#!/bin/bash

set -eu
source ./testing/tools/integration_runner.sh || exit 1

tarball="$1"
image_name="$2"

init_sg="$3"
src_cli="$4"

cmd_download="$5"
cmd_clear="$6"
cmd_upload="$7"
cmd_query="$8"

testdata_repos="$9"

port="7082"
url="http://localhost:$port"

SOURCEGRAPH_BASE_URL="$url"
export SOURCEGRAPH_BASE_URL

ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="true"
export ALLOW_SINGLE_DOCKER_CODE_INSIGHTS

run_server_image "$tarball" "$image_name" "$url" "$port"

echo '--- Initializing instance'
"$init_sg" initSG

echo '--- :horse: Running init-sg addRepos'
"$init_sg" addRepos -config "$testdata_repos"

echo '--- :brain: Running the test suite'

echo '--- :zero: downloading test data from GCS'
"$cmd_download"

echo '--- :one: clearing existing state'
"$cmd_clear"

echo '--- :two: integration test ./dev/codeintel-qa/cmd/upload'
"$cmd_upload" --timeout=5m --index-dir="./dev/codeintel-qa/testdata/indexes"

echo '--- :three: integration test ./dev/codeintel-qa/cmd/query'
"$cmd_query" --index-dir="./dev/codeintel-qa/testdata/indexes"

echo "--- "
