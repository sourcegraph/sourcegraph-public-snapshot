#!/usr/bin/env bash
# This runs a published or local server image for testing and development purposes.

IMAGE=${IMAGE:-sourcegraph/server:${TAG:-insiders}}
URL=${URL:-"http://localhost:7080"}
DATA=${DATA:-"/tmp/sourcegraph-data"}

echo "--- Checking for existing Sourcegraph instance at $URL"
if curl --output /dev/null --silent --head --fail "$URL"; then
  echo "❌ Can't run a new Sourcegraph instance on $URL because another instance is already running."
  echo "❌ The last time this happened, there was a runaway integration test run on the same Buildkite agent and the fix was to delete the pod and rebuild."
  exit 1
fi

# shellcheck disable=SC2153
case "$CLEAN" in
  "true")
    clean=y
    ;;
  "false")
    clean=n
    ;;
  *)
    echo -n "Do you want to delete $DATA and start clean? [Y/n] "
    read -r clean
    ;;
esac

if [ "$clean" != "n" ] && [ "$clean" != "N" ]; then
  echo "--- Deleting $DATA"
  rm -rf "$DATA"
fi

echo "--- Starting server ${IMAGE}"
docker run "$@" \
  --publish 7080:7080 \
  -e SRC_LOG_LEVEL=dbug \
  -e DEBUG=t \
  -e ALLOW_SINGLE_DOCKER_CODE_INSIGHTS \
  --volume "$DATA/config:/etc/sourcegraph" \
  --volume "$DATA/data:/var/opt/sourcegraph" \
  "$IMAGE"
