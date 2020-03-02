#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

if [ -z "$IMAGE" ]; then
    echo "Must specify \$IMAGE."
    exit 1
fi

URL="http://localhost:7080"

if curl --output /dev/null --silent --head --fail $URL; then
    echo "❌ Can't run a new Sourcegraph instance on $URL because another instance is already running."
    echo "❌ The last time this happened, there was a runaway e2e test run on the same Buildkite agent and the fix was to delete the pod and rebuild."
    exit 1
fi

# Switch the default Docker host to the dedicated e2e testing host
export DOCKER_HOST="$E2E_DOCKER_HOST"
export DOCKER_PASSWORD="$E2E_DOCKER_PASSWORD"
export DOCKER_USERNAME="$E2E_DOCKER_USERNAME"

echo "--- Copying $IMAGE to the dedicated e2e testing node..."
docker pull $IMAGE
echo "Copying $IMAGE to the dedicated e2e testing node... done"

echo "--- Running a daemonized $IMAGE as the test subject..."
CONTAINER="$(docker container run -d -e DEPLOY_TYPE=dev $IMAGE)"
trap 'kill $(jobs -p -r)'" ; docker logs --timestamps $CONTAINER ; docker container rm -f $CONTAINER ; docker image rm -f $IMAGE" EXIT

docker exec "$CONTAINER" apk add --no-cache socat
# Connect the server container's port 7080 to localhost:7080 so that e2e tests
# can hit it. This is similar to port-forwarding via SSH tunneling, but uses
# docker exec as the transport.
socat tcp-listen:7080,reuseaddr,fork system:"docker exec -i $CONTAINER socat stdio 'tcp:localhost:7080'" &

set +e
timeout 60s bash -c "until curl --output /dev/null --silent --head --fail $URL; do
    echo Waiting 5s for $URL...
    sleep 5
done"
if [ $? -ne 0 ]; then
    echo "^^^ +++"
    echo "$URL was not accessible within 60s. Here's the output of docker inspect and docker logs:"
    docker inspect "$CONTAINER"
    exit 1
fi
set -e
echo "Waiting for $URL... done"

echo "--- yarn"
# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network

echo "--- yarn run test-e2e"
pushd web
# `-pix_fmt yuv420p` makes a QuickTime-compatible mp4.
ffmpeg -y -f x11grab -video_size 1280x1024 -i "$DISPLAY" -pix_fmt yuv420p e2e.mp4 > ffmpeg.log 2>&1 &
env SOURCEGRAPH_BASE_URL="$URL" PERCY_ON=true ./node_modules/.bin/percy exec -- yarn run test-e2e
popd
