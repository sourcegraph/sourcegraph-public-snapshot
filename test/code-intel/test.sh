#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit
set -x

curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src

CONTAINER=sourcegraph-server

docker_logs() {
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
}

asdf install
yarn
yarn generate
pushd enterprise || exit
./cmd/server/pre-build.sh
./cmd/server/build.sh
docker run -d -p 7080:7080 --name "$CONTAINER" "$IMAGE"
popd || exit
trap docker_logs exit

sleep 15

pushd test/code-intel || exit
go run main.go
popd || exit

# shellcheck disable=SC1091
source /root/.profile

pushd internal/cmd/precise-code-intel-tester || exit
go build
./precise-code-intel-tester addrepos
./scripts/download.sh
./precise-code-intel-tester upload
sleep 10
./precise-code-intel-tester query
popd || exit
