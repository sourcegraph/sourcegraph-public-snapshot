#!/usr/bin/env bash

# This script runs the executors-e2e test suite against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

export TMP_DIR
TMP_DIR=$(mktemp -d)

cleanup() {
  pushd "$root_dir"/dev/ci/integration/executors/ 1>/dev/null
  docker-compose down --timeout 30 # seconds
  popd 1>/dev/null
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

export SERVER_IMAGE="us.gcr.io/sourcegraph-dev/server:${CANDIDATE_VERSION}"
export EXECUTOR_IMAGE="us.gcr.io/sourcegraph-dev/server:${CANDIDATE_VERSION}"
export EXECUTOR_FRONTEND_PASSWORD="hunter2hunter2hunter2"

echo "--- :terminal: Start server with executor"
cd "dev/ci/integration/executors"
export DATA="${TMP_DIR}"
docker-compose up -d

echo "--- :terminal: Wait for server to be up"
URL="http://localhost:7080"
timeout 120s bash -c "until curl --output /dev/null --silent --head --fail $URL; do
    echo Waiting 5s for $URL...
    sleep 5
done"

# ==========================

echo "--- :terminal: Run test.sh"
./test.sh
