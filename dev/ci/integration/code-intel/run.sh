#!/usr/bin/env bash

# This script runs the codeintel-qa test utility against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

echo "--- test.sh"
export IMAGE="us.gcr.io/sourcegraph-dev/server:${CANDIDATE_VERSION}"
./dev/ci/integration/run-integration.sh "${root_dir}/dev/ci/integration/code-intel/test.sh"
