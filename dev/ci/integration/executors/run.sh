#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

cleanup() {
  cd "$root_dir"/dev/ci/integration/executors/
  docker-compose down --timeout 30 # seconds
  cd "$root_dir"
}
trap cleanup EXIT

cd "dev/ci/integration/executors"
docker-compose up -d

# ==========================

echo "--- test.sh"
./test.sh
