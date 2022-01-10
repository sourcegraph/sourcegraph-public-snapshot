#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

set -ex

dev/ci/integration/setup-deps.sh
dev/ci/integration/setup-display.sh

cleanup() {
  cd "$root_dir"
  dev/ci/integration/cleanup-display.sh
  if [[ $(docker ps -aq | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rm -f $(docker ps -aq)
  fi
  if [[ $(docker images -q | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rmi -f $(docker images -q)
  fi
}
trap cleanup EXIT

# ==========================

echo "--- test.sh"
export IMAGE=${IMAGE:-"us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION"}
./dev/ci/integration/run-integration.sh "${root_dir}/dev/ci/integration/e2e/test.sh"
