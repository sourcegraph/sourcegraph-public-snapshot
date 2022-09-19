#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

dev/ci/integration/setup-deps.sh
dev/ci/integration/setup-display.sh

cleanup() {
  cd "$root_dir"
  dev/ci/integration/cleanup-display.sh
}
trap cleanup EXIT

# ==========================

echo "--- Running QA tests"

echo "--- test.sh"
export IMAGE=${IMAGE:-"us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION"}
set +x
# Hotfix (Owner: @jhchabran)
GITHUB_TOKEN="$(gcloud secrets versions access latest --secret=QA_GITHUB_TOKEN --quiet --project=sourcegraph-ci)"
export GITHUB_TOKEN
set -x
./dev/ci/integration/run-integration.sh "${root_dir}/dev/ci/integration/qa/test.sh"
