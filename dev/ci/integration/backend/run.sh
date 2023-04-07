#!/usr/bin/env bash

# This script runs the backend integration tests against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

echo "--- test.sh"

# Backend integration tests requires a GitHub Enterprise Token
set +x
# Hotfix (Owner: @mucles)
GITHUB_TOKEN="$(gcloud secrets versions access latest --secret=QA_GITHUB_TOKEN --quiet --project=sourcegraph-ci)"
export GITHUB_TOKEN
set -x
./dev/ci/integration/run-integration.sh "${root_dir}/dev/ci/integration/backend/test.sh"
