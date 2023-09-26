#!/usr/bin/env bash

set -eu
source ./testing/tools/integration_runner.sh || exit 1

tarball="$1"
image_name="$2"

gqltest="$3"
authtest="$4"

url="http://localhost:$PORT"

# Backend integration tests uses a specific GITHUB_TOKEN that is available as GHE_GITHUB_TOKEN
# because it refers to our internal GitHub enterprise instance used for testing.
if [ -n "${GHE_GITHUB_TOKEN:-''}" ]; then
  echo "GHE_GITHUB_TOKEN is empty"

  echo "--- debug"
  echo "$(env | grep GITHUB)"
  GITHUB_TOKEN=""
else
  GITHUB_TOKEN="$GHE_GITHUB_TOKEN"
fi
export GITHUB_TOKEN

ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="true"
export ALLOW_SINGLE_DOCKER_CODE_INSIGHTS

run_server_image "$tarball" "$image_name" "$url" "$PORT"

echo "--- integration test ./dev/gqltest -long"
"$gqltest" -long -base-url "$url"

echo "--- sleep 5s to wait for site configuration to be restored from gqltest"
sleep 5

echo "--- integration test ./dev/authtest -long"
"$authtest" -long -base-url "$url" -email "gqltest@sourcegraph.com" -username "gqltest-admin"

echo "--- done"
