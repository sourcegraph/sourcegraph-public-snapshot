#!/usr/bin/env bash

set -ex

echo "Running a daemonized sourcegraph/server as the test subject..."
SERVER_CONTAINER_ID="$(docker container run --rm -d -p 7080:7080 sourcegraph/server:3.1.1)"
trap "docker container stop $SERVER_CONTAINER_ID" EXIT
SERVER_URL="$(docker container port "$SERVER_CONTAINER_ID")"
set +e
until curl --output /dev/null --silent --head --fail "$SERVER_URL"; do
    echo "Waiting 1s for $SERVER_URL..."
    sleep 1
done
set -e
echo "Waiting for $SERVER_URL... done"

export FORCE_COLOR="1"
export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=""
yarn --frozen-lockfile --network-timeout 60000

pushd web
env SOURCEGRAPH_BASE_URL="$SERVER_URL" yarn run test-e2e -t 'theme'
popd
