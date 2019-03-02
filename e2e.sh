#!/usr/bin/env bash

set -ex

trap 'kill $(jobs -p)' EXIT

echo "Running a daemonized sourcegraph/server as the test subject..."
SERVER_CONTAINER_ID="$(docker container run --rm -d -p :7080 sourcegraph/server:3.1.1)"
trap "docker container stop $SERVER_CONTAINER_ID" EXIT

SERVER_HOSTPORT="$(docker container port "$SERVER_CONTAINER_ID" 7080)"

docker exec "$SERVER_CONTAINER_ID" apk add --no-cache socat
apt-get install -y socat
socat tcp-listen:7080,reuseaddr,fork system:"docker exec -i $SERVER_CONTAINER_ID socat stdio 'tcp:$SERVER_HOSTPORT'" &

set +e
until curl --output /dev/null --silent --head --fail "$SERVER_HOSTPORT"; do
    echo "Waiting 1s for $SERVER_HOSTPORT..."
    sleep 1
done
set -e
echo "Waiting for $SERVER_HOSTPORT... done"

export FORCE_COLOR="1"
export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=""
yarn --frozen-lockfile --network-timeout 60000

pushd web
env SOURCEGRAPH_BASE_URL="http://$SERVER_HOSTPORT" yarn run test-e2e -t 'theme'
popd
