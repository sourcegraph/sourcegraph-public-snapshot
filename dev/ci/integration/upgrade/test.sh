#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -e

URL="${1:-"http://localhost:7080"}"

# In CI, provide a directory and container name unique to this job
IDENT=${BUILDKITE_JOB_ID:-$(openssl rand -hex 12)}
export DATA="/tmp/sourcegraph-data-${IDENT}"

cleanup() {
  echo "--- dump server logs"
  docker logs --timestamps "$CONTAINER" >"$root_dir/$CONTAINER.log" 2>&1

  echo "--- Deleting $DATA"
  rm -rf "$DATA"
}

trap cleanup EXIT

# Run and initialize an old Sourcegraph release
echo "--- start sourcegraph $MINIMUM_UPGRADEABLE_VERSION"
CONTAINER="sourcegraph-old-${IDENT}"
IMAGE=sourcegraph/server:$MINIMUM_UPGRADEABLE_VERSION CLEAN="true" ./dev/run-server-image.sh -d --name "$CONTAINER"
sleep 15
pushd internal/cmd/init-sg
go build
./init-sg initSG
popd
# shellcheck disable=SC1091
source /root/.sg_envrc

SOURCEGRAPH_REPORTED_VERSION_OLD=$(curl -fs "$URL/__version")
echo
echo "--- Sourcegraph instance (before upgrade) is reporting version: '$SOURCEGRAPH_REPORTED_VERSION_OLD'"

# Stop old Sourcegraph release
docker container stop "$CONTAINER"
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
echo "--- start candidate"
CONTAINER="sourcegraph-new-${IDENT}"
IMAGE=us.gcr.io/sourcegraph-dev/server:bazel-${CANDIDATE_VERSION} CLEAN="false" ./dev/run-server-image.sh -d --name "$CONTAINER"
sleep 15

# Run tests
echo "--- TEST: Checking Sourcegraph instance is accessible"
curl -f "$URL"
curl -f "$URL"/healthz

SOURCEGRAPH_REPORTED_VERSION_NEW=$(curl -fs "$URL/__version")
echo
echo "--- Sourcegraph instance (after upgrade) is reporting version: '$SOURCEGRAPH_REPORTED_VERSION_NEW'"

if [ "$SOURCEGRAPH_REPORTED_VERSION_NEW" == "$SOURCEGRAPH_REPORTED_VERSION_OLD" ]; then
  echo "Error: Instance version unchanged after upgrade" 1>&2
  exit 1
fi

echo "--- TEST: Running tests"
pushd client/web
pnpm run test:regression:core
popd
