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

pushd enterprise
./cmd/server/pre-build.sh
./cmd/server/build.sh
popd

echo "TEST: Running E2E tests"
./dev/ci/e2e.sh
docker image rm -f "${IMAGE}"
