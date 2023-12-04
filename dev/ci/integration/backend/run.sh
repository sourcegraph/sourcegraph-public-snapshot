#!/usr/bin/env bash

# This script runs the backend integration tests against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

echo "--- test.sh"

# Backend integration tests requires a GitHub Enterprise Token
GITHUB_TOKEN=$GHE_GITHUB_TOKEN
GITHUB_TOKEN=$GITHUB_TOKEN ./dev/ci/integration/run-integration.sh "${root_dir}/dev/ci/integration/backend/test.sh"
