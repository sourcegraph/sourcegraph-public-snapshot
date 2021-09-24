#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
root_dir="$(dirname "${BASH_SOURCE[0]}")/../../../.."
cd "$root_dir"
root_dir=$(pwd)

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
pushd internal/cmd/init-sg
go build
./init-sg initSG
popd
# Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
# shellcheck disable=SC1091
source /root/.profile
set -x

# Stop old Sourcegraph release
docker container stop sourcegraph-old
sleep 5

# Migrate DB if on version < 3.27.0
regex="3\.26\.[0-9]"
OLD=11
NEW=12
SRC_DIR=/tmp/sourcegraph
if [[ $MINIMUM_UPGRADEABLE_VERSION =~ $regex ]]; then
  docker run \
    -w /tmp/upgrade \
    -v "$SRC_DIR/data/postgres-$NEW-upgrade:/tmp/upgrade" \
    -v "$SRC_DIR/data/postgresql:/var/lib/postgresql/$OLD/data" \
    -v "$SRC_DIR/data/postgresql-$NEW:/var/lib/postgresql/$NEW/data" \
    "tianon/postgres-upgrade:$OLD-to-$NEW"

  mv "$SRC_DIR/data/"{postgresql,postgresql-$OLD}
  mv "$SRC_DIR/data/"{postgresql-$NEW,postgresql}

  curl -fsSL -o "$SRC_DIR/data/postgres-$NEW-upgrade/optimize.sh" https://raw.githubusercontent.com/sourcegraph/sourcegraph/master/cmd/server/rootfs/postgres-optimize.sh

  docker run \
    --entrypoint "/bin/bash" \
    -w /tmp/upgrade \
    -v "$SRC_DIR/data/postgres-$NEW-upgrade:/tmp/upgrade" \
    -v "$SRC_DIR/data/postgresql:/var/lib/postgresql/data" \
    "postgres:$NEW" \
    -c 'chown -R postgres $PGDATA . && gosu postgres bash ./optimize.sh $PGDATA'
fi

# Upgrade to current candidate image. Capture logs for the attempted upgrade.
CONTAINER=sourcegraph-new
docker_logs() {
  pushd "$root_dir"
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
  popd
}
IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION CLEAN="false" ./dev/run-server-image.sh -d --name $CONTAINER
trap docker_logs exit
sleep 15

# Run tests
echo "TEST: Checking Sourcegraph instance is accessible"
curl -f http://localhost:7080
curl -f http://localhost:7080/healthz
echo "TEST: Running tests"
echo "TEST: Downloading Puppeteer"
yarn --cwd client/shared run download-puppeteer-browser
pushd client/web
yarn run test:regression:core
popd
