#!/usr/bin/env bash

# shellcheck disable=SC1091
# source /root/.profile
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

URL="http://localhost:7080"
timeout 120s bash -c "until curl --output /dev/null --silent --head --fail $URL; do
    echo Waiting 5s for $URL...
    sleep 5
done"

# ==========================

echo "--- test.sh"
./test.sh
