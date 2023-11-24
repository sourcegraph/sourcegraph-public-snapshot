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
  docker-compose logs >"${root_dir}/docker-compose.log"
  docker-compose down --volumes --timeout 30 # seconds
  docker volume rm executors-e2e || true
  popd 1>/dev/null
  rm -rf "${TMP_WORK_DIR}"
}
trap cleanup EXIT

# registry="us.gcr.io/sourcegraph-dev/"
registry="us-central1-docker.pkg.dev/sourcegraph-ci/rfc795-internal"
export POSTGRES_IMAGE="$registry/postgres-12-alpine:${CANDIDATE_VERSION}"
export SERVER_IMAGE="$registry/server:${CANDIDATE_VERSION}"
export EXECUTOR_IMAGE="$registry/executor:${CANDIDATE_VERSION}"
export EXECUTOR_FRONTEND_PASSWORD="hunter2hunter2hunter2"
export SOURCEGRAPH_LICENSE_GENERATION_KEY="${SOURCEGRAPH_LICENSE_GENERATION_KEY:-""}"
export TMP_DIR
export DATA
if [ -n "${DOCKER_GATEWAY_HOST}" ]; then
  DOCKER_HOST="tcp://${DOCKER_GATEWAY_HOST:-host.docker.internal}:2375"
  export DOCKER_HOST
fi

# Need to pull this image pre-execution as the docker executor doesn't have a
# credential to pull this image.
BATCHESHELPER_IMAGE="$registry/batcheshelper:${CANDIDATE_VERSION}"
docker pull "${BATCHESHELPER_IMAGE}"

echo "--- :terminal: Start server with executor"
pushd "dev/ci/integration/executors" 1>/dev/null

# Temporary workaround, see https://github.com/sourcegraph/sourcegraph/issues/44816
envsubst >"${TMP_WORK_DIR}/site-config.json" <./tester/config/site-config.json
docker volume create executors-e2e
docker container create --name temp -v executors-e2e:/data busybox
docker cp "${TMP_WORK_DIR}/site-config.json" temp:/data
rm "${TMP_WORK_DIR}/site-config.json"
docker rm temp

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
