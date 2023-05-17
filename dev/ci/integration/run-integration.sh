#!/usr/bin/env bash

# This script sets up a Sourcegraph instance for integration testing. This script expects to be
# passed a path to a bash script that runs the actual tests against a running instance. The passed
# script will be passed a single parameter: the target URL from which the instance is accessible.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
root_dir=$(pwd)
set -ex

if [ -z "$IMAGE" ]; then
  echo "Must specify \$IMAGE."
  exit 1
fi

URL="http://localhost:7080"

# In CI, provide a directory and container name unique to this job
IDENT=${BUILDKITE_JOB_ID:-$(openssl rand -hex 12)}
export DATA="/tmp/sourcegraph-data-${IDENT}"
export CONTAINER="sourcegraph-${IDENT}"

function docker_cleanup() {
  echo "--- docker cleanup"
  if [[ $(docker ps -aq | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rm -f $(docker ps -aq)
  fi
  if [[ $(docker images -q | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rmi -f $(docker images -q)
  fi
  docker volume prune -f

  echo "--- Deleting $DATA"
  rm -rf "$DATA"
}

# Do a pre-run cleanup
docker_cleanup

function cleanup() {
  exit_status=$?
  if [ $exit_status -ne 0 ]; then
    # Expand the output if our run failed.
    echo "^^^ +++"
  fi

  echo "--- dump server logs"
  docker logs --timestamps "$CONTAINER" >"$root_dir/server.log" 2>&1

  echo "--- $CONTAINER cleanup"
  docker container rm -f "$CONTAINER"
  docker image rm -f "$IMAGE"

  docker_cleanup

  if [ $exit_status -ne 0 ]; then
    # This command will fail, so our last step will be expanded. We don't want
    # to expand "docker cleanup" so we add in a dummy section.
    echo "--- integration test failed"
    echo "See integration test section for test runner logs, and uploaded artifacts for server logs."
  fi
}
trap cleanup EXIT

echo "--- Running a daemonized $IMAGE as the test subject..."
CLEAN="true" ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="true" "${root_dir}"/dev/run-server-image.sh -d --name "$CONTAINER"

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
