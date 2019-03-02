#!/usr/bin/env bash

set -ex

echo "Running a daemonized sourcegraph/server as the test subject..."
CONTAINER="$(docker container run --rm -d sourcegraph/server:insiders)"
trap 'kill $(jobs -p)'" ; docker container stop $CONTAINER" EXIT

# hax
docker exec "$CONTAINER" apk add --no-cache socat
apt-get install -y socat
socat tcp-listen:7080,reuseaddr,fork system:"docker exec -i $CONTAINER socat stdio 'tcp:localhost:7080'" &

URL="http://localhost:7080"

timeout 30s bash -c "until curl --output /dev/null --silent --head --fail $URL; do
    echo Waiting 5s for $URL...
    sleep 5
done"
echo "Waiting for $URL... done"

export FORCE_COLOR="1"
export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=""
yarn --frozen-lockfile --network-timeout 60000

pushd web
env SOURCEGRAPH_BASE_URL="$URL" yarn run test-e2e -t 'theme'
popd
