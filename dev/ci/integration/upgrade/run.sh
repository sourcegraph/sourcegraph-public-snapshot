#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
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

echo "--- test.sh"
"${root_dir}"/dev/ci/integration/upgrade/test.sh
