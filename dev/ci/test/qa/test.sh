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

CONTAINER=sourcegraph-server

docker_logs() {
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
}

if [[ $VAGRANT_RUN_ENV = "CI" ]]; then
  IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION
else
  # shellcheck disable=SC2034
  IMAGE=sourcegraph/server:insiders
fi

./dev/run-server-image.sh -d --name $CONTAINER
trap docker_logs exit
sleep 15

pushd internal/cmd/init-sg
go build
./init-sg initSG
popd
# Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
# shellcheck disable=SC1091
source /root/.profile
set -x

echo "TEST: Checking Sourcegraph instance is accessible"
curl -f http://localhost:7080
curl -f http://localhost:7080/healthz
echo "TEST: Running tests"
# Run all tests, and error if one fails
test_status=0
pushd client/web
yarn run test:regression || test_status=1
popd
exit $test_status
