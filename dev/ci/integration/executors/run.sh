#!/usr/bin/env bash

# This script runs the executors-e2e test suite against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

TMP_WORK_DIR=$(mktemp -d)
TMP_DIR="${TMP_WORK_DIR}/executor-tmp"
mkdir "${TMP_DIR}"
DATA="${TMP_WORK_DIR}/data"
mkdir "${DATA}"
mkdir "${DATA}/data"
mkdir "${DATA}/config"

cleanup() {
  pushd "$root_dir"/dev/ci/integration/executors/ 1>/dev/null
  docker-compose logs >"${root_dir}/logs.txt"
  docker-compose down --timeout 30 # seconds
  popd 1>/dev/null
  rm -rf "${TMP_WORK_DIR}"
}
trap cleanup EXIT

export SERVER_IMAGE="us.gcr.io/sourcegraph-dev/server:${CANDIDATE_VERSION}"
export EXECUTOR_IMAGE="us.gcr.io/sourcegraph-dev/executor:${CANDIDATE_VERSION}"
export EXECUTOR_FRONTEND_PASSWORD="hunter2hunter2hunter2"
export SOURCEGRAPH_LICENSE_GENERATION_KEY="${SOURCEGRAPH_LICENSE_GENERATION_KEY:-""}"
export TMP_DIR
export DATA

echo "--- :terminal: Start server with executor"
pushd "dev/ci/integration/executors" 1>/dev/null
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
