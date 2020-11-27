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

# Run and initialize an old Sourcegraph release
IMAGE=sourcegraph/server:$MINIMUM_UPGRADEABLE_VERSION ./dev/run-server-image.sh -d --name sourcegraph-old
sleep 15
go run dev/ci/test/init-server.go

# Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
# shellcheck disable=SC1091
source /root/.profile
set -x

# Stop old Sourcegraph release
docker container stop sourcegraph-old
sleep 5

# Upgrade to current candidate image. Capture logs for the attempted upgrade.
CONTAINER=sourcegraph-new
docker_logs() {
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
}
IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION CLEAN="false" ./dev/run-server-image.sh -d --name $CONTAINER
trap docker_logs exit
sleep 15

# Run tests
echo "TEST: Checking Sourcegraph instance is accessible"
curl -f http://localhost:7080
curl -f http://localhost:7080/healthz
echo "TEST: Running tests"
pushd client/web
yarn run test:regression:core
popd
