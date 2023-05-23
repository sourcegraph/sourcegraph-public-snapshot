#!/bin/bash

set -eu
source ./testing/tools/integration_runner.sh || exit 1

tarball="$1"
image_name="$2"

e2e_test="$3"
e2e_test_target="$4"
e2e_test_data_dir=$(bazel info bazel-test "$e2e_test_target")

url="http://localhost:7080"

# Backend integration tests uses a specific GITHUB_TOKEN that is available as GHE_GITHUB_TOKEN
# because it refers to our internal GitHub enterprise instance used for testing.
GITHUB_TOKEN="$GHE_GITHUB_TOKEN"
export GITHUB_TOKEN

ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="true"
export ALLOW_SINGLE_DOCKER_CODE_INSIGHTS

run_server_image "$tarball" "$image_name" "$url" "7080"

echo "--- e2e test //client/web/src/end-to-end:e2e"
"$e2e_test" --input_file="$e2e_test_data_dir"/input.txt --output_file="$e2e_test_data_dir"/output.txt

echo "--- "

