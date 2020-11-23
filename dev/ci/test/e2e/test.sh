#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
root_dir="$(dirname "${BASH_SOURCE[0]}")/../../../.."
cd "$root_dir"

set -ex

dev/ci/test/setup-deps.sh
dev/ci/test/setup-display.sh

cleanup() {
  cd "$root_dir"
  dev/ci/test/cleanup-display.sh
}
trap cleanup EXIT

# ==========================

echo "TEST: Running E2E tests"
IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION ./dev/ci/e2e.sh
