#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit
set -x

asdf install
yarn generate

cd ./client/web || exit

# Run and initialize an old Sourcegraph release
docker run --name sourcegraph-old --detach --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm --volume ~/.sourcegraph/config:/etc/sourcegraph --volume ~/.sourcegraph/data:/var/opt/sourcegraph \
  sourcegraph/server:"$TEST_UPGRADE_FROM_SOURCEGRAPH_VERSION"
E2E_INIT=true SOURCEGRAPH_BASE_URL=http://localhost:7080 yarn run test:regression -t 'Initialize new Sourcegraph instance'
# Upgrade to current candidate image
docker container stop sourcegraph-old
docker run --name sourcegraph-new --detach --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm --volume ~/.sourcegraph/config:/etc/sourcegraph --volume ~/.sourcegraph/data:/var/opt/sourcegraph \
  sourcegraph/server:"$TEST_UPGRADE_TO_SOURCEGRAPH_VERSION"

# Run tests
echo "TEST: Running regression tests"
SOURCEGRAPH_BASE_URL=http://localhost:7080 yarn run test:regression
echo "TEST: Checking Sourcegraph instance is accessible"
curl -f http://localhost:3370
curl -f http://localhost:3370/healthz
