#!/usr/bin/env bash

# This script sets up a Sourcegraph instance for integration testing. This script expects to be
# passed a path to a bash script that runs the actual tests against a running instance. The passed
# script will be passed a single parameter: the target URL from which the instance is accessible.

cd "$(dirname "${BASH_SOURCE[0]}")/../../"
root_dir=$(pwd)
set -ex

if [ -z "$IMAGE" ]; then
  echo "Must specify \$IMAGE."
  exit 1
fi

URL="http://localhost:7080"

if curl --output /dev/null --silent --head --fail $URL; then
  echo "❌ Can't run a new Sourcegraph instance on $URL because another instance is already running."
  echo "❌ The last time this happened, there was a runaway integration test run on the same Buildkite agent and the fix was to delete the pod and rebuild."
  exit 1
fi

echo "--- Running a daemonized $IMAGE as the test subject..."
CONTAINER="$(docker container run -d -e GOTRACEBACK=all "$IMAGE")"
function cleanup() {
  exit_status=$?
  if [ $exit_status -ne 0 ]; then
    # Expand the output if our run failed.
    echo "^^^ +++"
  fi

  jobs -p -r | xargs kill
  echo "--- dump server logs"
  docker logs --timestamps "$CONTAINER" >"$root_dir/server.log" 2>&1
  echo "--- docker cleanup"
  docker container rm -f "$CONTAINER"
  docker image rm -f "$IMAGE"

  if [ $exit_status -ne 0 ]; then
    # This command will fail, so our last step will be expanded. We don't want
    # to expand "docker cleanup" so we add in a dummy section.
    echo "--- integration test failed"
    echo "See integration test section for test runner logs, and uploaded artefacts for server logs."
  fi
}
trap cleanup EXIT

docker exec "$CONTAINER" apk add --no-cache socat
# Connect the server container's port 7080 to localhost:7080 so that integration tests
# can hit it. This is similar to port-forwarding via SSH tunneling, but uses `docker exec`
# as the transport.
socat tcp-listen:7080,reuseaddr,fork system:"docker exec -i $CONTAINER socat stdio 'tcp:localhost:7080'" &

echo "--- Waiting for $URL to be up"
set +e
timeout 120s bash -c "until curl --output /dev/null --silent --head --fail $URL; do
    echo Waiting 5s for $URL...
    sleep 5
done"
# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
  echo "^^^ +++"
  echo "$URL was not accessible within 120s."
  docker inspect "$CONTAINER"
  exit 1
fi
set -e
echo "Waiting for $URL... done"

# Run tests against instance
"${1}" "${URL}"
