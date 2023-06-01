#!/bin/bash

set -eu
source ./testing/tools/integration_runner.sh || exit 1

tarball="$1"
image_name="$2"

gqltest="$3"
authtest="$4"

port="7081"
url="http://localhost:$port"

# Backend integration tests uses a specific GITHUB_TOKEN that is available as GHE_GITHUB_TOKEN
# because it refers to our internal GitHub enterprise instance used for testing.
GITHUB_TOKEN="$GHE_GITHUB_TOKEN"
export GITHUB_TOKEN

ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="true"
export ALLOW_SINGLE_DOCKER_CODE_INSIGHTS

run_server_image "$tarball" "$image_name" "$url" "$port"

echo "--- integration test ./dev/gqltest -long"
"$gqltest" -long -base-url "$url"

echo "--- sleep 5s to wait for site configuration to be restored from gqltest"
sleep 5

echo "--- integration test ./dev/authtest -long"
"$authtest" -long -base-url "$url" -email "gqltest@sourcegraph.com" -username "gqltest-admin"

echo "--- done"
