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
go build -o /usr/local/bin/init-sg
popd

pushd dev/ci/test/code-intel
init-sg initSG
# # Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
source /root/.profile
set -x
init-sg addRepos -config repos.json
popd

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
