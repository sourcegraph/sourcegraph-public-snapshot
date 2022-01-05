#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
root_dir="$(dirname "${BASH_SOURCE[0]}")/../../../.."
cd "$root_dir"
root_dir=$(pwd)

set -ex

dev/ci/test/setup-deps.sh
dev/ci/test/setup-display.sh

# ==========================

docker_logs() {
  echo "--- dump server logs"
  docker logs --timestamps "$CONTAINER" >"$root_dir/$CONTAINER.log" 2>&1
}

cleanup() {
  docker_logs
  cd "$root_dir"
  docker rm -f "$CONTAINER"
  if [[ $(docker images -q | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rmi -f $(docker images -q)
  fi

}

if [[ $BUILDKITE = "true" ]]; then
  IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION
else
  # shellcheck disable=SC2034
  IMAGE=sourcegraph/server:insiders
fi

CONTAINER=sourcegraph-server
CLEAN="true" ./dev/run-server-image.sh -d --name $CONTAINER
trap cleanup EXIT
sleep 15

echo "--- init sourcegraph"
pushd internal/cmd/init-sg
go build
./init-sg initSG
popd
# Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
# shellcheck disable=SC1091
source /root/.sg_envrc
set -x

echo "--- TEST: Checking Sourcegraph instance is accessible"
curl -f http://localhost:7080
curl -f http://localhost:7080/healthz
echo "--- TEST: Running tests"
# Run all tests, and error if one fails
test_status=0
pushd client/web
yarn run test:regression || test_status=1
popd
exit $test_status
