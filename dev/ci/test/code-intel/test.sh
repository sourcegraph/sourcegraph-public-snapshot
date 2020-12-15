#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
root_dir="$(dirname "${BASH_SOURCE[0]}")/../../../.."
cd "$root_dir"

set -ex

dev/ci/test/setup-deps.sh

# ==========================

CONTAINER=sourcegraph-server

docker_logs() {
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
}

if [[ $VAGRANT_RUN_ENV="CI" ]]; then
  IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION
else
  IMAGE=sourcegraph/server:latest
fi

./dev/run-server-image.sh -d --name $CONTAINER
trap docker_logs exit
sleep 15

pushd internal/cmd/init-server
go build
./init-server "$root_dir/dev/ci/test/code-intel/extsvc.json"
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
pushd internal/cmd/precise-code-intel-tester
go build
./scripts/download.sh
./precise-code-intel-tester upload
sleep 10
./precise-code-intel-tester query
popd
