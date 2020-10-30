#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit
set -x

asdf install
yarn install
yarn generate

# shellcheck disable=SC1091
source /root/.profile

cd ./client/web || exit

# Run and initialize an old Sourcegraph release
IMAGE=sourcegraph/server:$TEST_UPGRADE_FROM_SOURCEGRAPH_VERSION ../../dev/run-server-image.sh -d --name sourcegraph-old
sleep 15
E2E_INIT=true SOURCEGRAPH_BASE_URL=http://localhost:7080 yarn run test:regression -t 'Initialize new Sourcegraph instance'

# Upgrade to current candidate image
docker container stop sourcegraph-old
sleep 5
IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION ../../dev/run-server-image.sh -d --name sourcegraph-new
sleep 15

# Run tests
echo "TEST: Running regression tests"
SOURCEGRAPH_BASE_URL=http://localhost:7080 yarn run test:regression
echo "TEST: Checking Sourcegraph instance is accessible"
curl -f http://localhost:3370
curl -f http://localhost:3370/healthz
