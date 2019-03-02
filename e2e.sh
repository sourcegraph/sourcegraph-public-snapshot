#!/usr/bin/env bash

set -ex

trap 'kill $(jobs -p)' EXIT

echo "Running a daemonized sourcegraph/server as the test subject..."
SERVER_CONTAINER_ID="$(docker container run --rm -d -p :7080 sourcegraph/server:3.1.1)"
trap "docker container stop $SERVER_CONTAINER_ID" EXIT

docker exec "$SERVER_CONTAINER_ID" apk add --no-cache socat
apt-get install -y socat
socat tcp-listen:5000,reuseaddr,fork system:"docker exec -i 8a0956b1b1ef socat stdio 'tcp:localhost:7080'" &

SERVER_URL="$(docker container port "$SERVER_CONTAINER_ID" 7080)"
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
