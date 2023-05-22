#!/bin/bash

set -eu

# Grabbing tools and args
before_run_server_image="$1"
run_server_image="$2"
server_image="$3"
gqltest="$4"
authtest="$5"

# Settings
URL="http://localhost:7080"
IDENT=${BUILDKITE_JOB_ID:-$(openssl rand -hex 12)}
CONTAINER="integration-backend-$IDENT"
TIMEOUT=60
IMAGE="server:candidate"
GITHUB_TOKEN="$GHE_GITHUB_TOKEN"
export GITHUB_TOKEN

# Ensure we exit with a clean slate regardless of the outcome
function cleanup() {
  exit_status=$?
  if [ $exit_status -ne 0 ]; then
    # Expand the output if our run failed.
    echo "^^^ +++"
  fi

  echo "--- dump server logs"
  docker logs --timestamps "$CONTAINER"
  echo "--- "

  echo "--- $CONTAINER cleanup"
  docker container rm -f "$CONTAINER"
  docker image rm -f "$IMAGE"

  if [ $exit_status -ne 0 ]; then
    # This command will fail, so our last step will be expanded. We don't want
    # to expand "docker cleanup" so we add in a dummy section.
    echo "--- integration test failed"
    echo "See integration test section for test runner logs, and uploaded artifacts for server logs."
  fi
}
trap cleanup EXIT

# Ensuring clean slate before running image
"$before_run_server_image"

echo "--- Running a daemonized server:candidate as the test subject..."
"$run_server_image" "$server_image" server:candidate --rm -d --name "$CONTAINER"

echo "--- Waiting for $URL to be up"
set +e

t=1
# timeout is a coreutils extension, not available to us here
curl --output /dev/null --silent --head --fail $URL
# shellcheck disable=SC2181
while [ ! $? -eq 0 ]; do
  sleep 5
  t=$(( $t + 5 ))
  if [ "$t" -gt $TIMEOUT ]; then
  echo "$URL was not accessible within $TIMEOUT."
  docker inspect "$CONTAINER"
    exit 1
  fi

  curl --output /dev/null --silent --head --fail $URL
done
set -e

echo '--- integration test ./dev/gqltest -long'
"$gqltest" -long -base-url "$URL"

echo '--- sleep 5s to wait for site configuration to be restored from gqltest'
sleep 5

echo '--- integration test ./dev/authtest -long'
"$authtest" -long -base-url "$URL" -email "gqltest@sourcegraph.com" -username "gqltest-admin"
