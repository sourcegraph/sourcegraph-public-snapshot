#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit

set -x

test/setup-deps.sh

# ==========================

CONTAINER=sourcegraph-server

docker_logs() {
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
}

pushd enterprise || exit
./cmd/server/pre-build.sh
./cmd/server/build.sh
docker run -d -p 7080:7080 --name "$CONTAINER" "$IMAGE"
popd || exit
trap docker_logs exit
sleep 15

go run test/init-server.go

# Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
# shellcheck disable=SC1091
source /root/.profile
set -x

pushd internal/cmd/precise-code-intel-tester || exit
go build
./precise-code-intel-tester addrepos
./scripts/download.sh
./precise-code-intel-tester upload
sleep 10
./precise-code-intel-tester query
popd || exit

# ==========================
