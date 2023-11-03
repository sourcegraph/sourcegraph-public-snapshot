#!/usr/bin/env bash

set -e

DB_STARTUP_TIMEOUT="${DB_STARTUP_TIMEOUT:-10s}"

function must_be_CI() {
  if [ "${BUILDKITE:-}" != "true" ]; then
    echo "⚠️ This script is NOT running on a Buildkite agent."
    echo "👉 Aborting."
    exit 1
  fi
}

function ensure_clean_slate() {
  echo "--- Ensuring clean slate before running server container"

  local running
  running=$(docker ps -aq | wc -l)
  if [[ "$running" -gt 0 ]]; then
    echo "⚠️ Found $running running containers, deleting them."
    # shellcheck disable=SC2046
    docker rm -f $(docker ps -aq)
  else
    echo "Found 0 running containers."
  fi

  local images
  images=$(docker images -q | wc -l)
  if [[ "$images" -gt 0 ]]; then
    echo "⚠️ Found $images images, deleting them."
    # shellcheck disable=SC2046
    docker rmi -f $(docker images -q)
  else
    echo "Found 0 images."
  fi

  echo "--- done"
}

function is_present() {
  if [ -n "$1" ]; then
    echo "present"
  else
    echo "blank"
  fi
}

function must_not_be_running() {
  local url
  url="$1"
  if curl --output /dev/null --silent --head --fail "$url"; then
    echo "❌ Can't run a new server instance on $url because another instance is already running."
    exit 1
  fi
}

function generate_unique_container_name() {
  local prefix="$1"
  prefix="$1"
  local ident
  ident="$(openssl rand -hex 12)"
  echo "$prefix-$ident"
}

function _run_server_image() {
  if [ -z "$DB_STARTUP_TIMEOUT" ]; then
    echo "❌ DB_STARTUP_TIMEOUT must be defined"
  fi
  if [ -z "$SOURCEGRAPH_LICENSE_GENERATION_KEY" ]; then
    echo "❌ SOURCEGRAPH_LICENSE_GENERATION_KEY must be defined"
  fi
  if [ -z "$SOURCEGRAPH_LICENSE_KEY" ]; then
    echo "❌ SOURCEGRAPH_LICENSE_KEY must be defined"
  fi

  local image_tarball
  image_tarball="$1"
  local image_name
  image_name="$2"
  local url
  url="$3"
  local server_port
  server_port="$4"
  local container_name
  container_name="$5"
  local docker_args
  # shellcheck disable=SC2124
  docker_args="${@:6}"

  local guid
  local uid
  uid="$(id -u)"
  gid="$(id -g)"

  echo "--- Loading server image"
  echo "Loading $image_tarball in Docker"
  "$image_tarball" # this is a shell script

  echo "-- Starting $image_name"
  echo "Listening at: $url"
  echo "Database startup timeout: $DB_STARTUP_TIMEOUT"
  echo "License key generator present: $(is_present "$SOURCEGRAPH_LICENSE_GENERATION_KEY")"
  echo "License key present: $(is_present "$SOURCEGRAPH_LICENSE_GENERATION_KEY")"

  echo "Allow single docker image code insights: $ALLOW_SINGLE_DOCKER_CODE_INSIGHTS"
  echo "GRPC Feature flag: $SG_FEATURE_FLAG_GRPC"

  mkdir -p "$data/config"
  mkdir -p "$data/data"

  # shellcheck disable=SC2086
  docker run $docker_args \
    -d \
    --name "$container_name" \
    --publish "$server_port":7080 \
    --platform linux/amd64 \
    -u "$uid:$gid" \
    -e BAZEL_SKIP_OOB_INFER_VERSION=true \
    -e ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="$ALLOW_SINGLE_DOCKER_CODE_INSIGHTS" \
    -e SOURCEGRAPH_LICENSE_GENERATION_KEY="$SOURCEGRAPH_LICENSE_GENERATION_KEY" \
    -e SG_FEATURE_FLAG_GRPC="$SG_FEATURE_FLAG_GRPC" \
    -e DB_STARTUP_TIMEOUT="$DB_STARTUP_TIMEOUT" \
    "$image_name"

  echo "-- Listening at $url"
}

function wait_until_container_ready() {
  local name
  name="$1"
  local url
  url="$2"
  local timeout
  timeout="$3"

  echo "--- Waiting for $url to be up"
  set +e

  t=1
  # timeout is a coreutils extension, not available to us here
  curl --output /dev/null --silent --head --fail "$url"
  # shellcheck disable=SC2181
  while [ ! $? -eq 0 ]; do
    sleep 5
    t=$(( t + 5 ))
    if [ "$t" -gt "$timeout" ]; then
      echo "$url was not accessible within $timeout."
      docker inspect "$name"
      exit 1
    fi

    curl --output /dev/null --silent --head --fail "$url"
  done
  set -e
}

function run_server_image() {
  local image_tarball
  image_tarball="$1"
  local image_name
  image_name="$2"
  local url
  url="$3"
  local server_port
  server_port="$4"

  must_not_be_running "$url"

  local container_name
  container_name=$(generate_unique_container_name "server-integration")

  # we want those to be expanded right now, on purpose.
  # shellcheck disable=SC2064
  trap "cleanup $image_name $container_name" EXIT
  _run_server_image "$image_tarball" "$image_name" "$url" "$server_port" "$container_name"

  wait_until_container_ready "$container_name" "$url" 60
}

# Ensure we exit with a clean slate regardless of the outcome
function cleanup() {
  exit_status=$?

  local image
  image="$1"
  local container
  container="$2"

  if [ $exit_status -ne 0 ]; then
    # Expand the output if our run failed.
    echo "^^^ +++"
  fi

  echo "--- dump server logs"
  docker logs --timestamps "$container"
  echo "--- done"

  echo "--- $container cleanup"
  docker container rm -f "$container"
  docker image rm -f "$image"

  if [ $exit_status -ne 0 ]; then
    # This command will fail, so our last step will be expanded. We don't want
    # to expand "docker cleanup" so we add in a dummy section.
    echo "--- integration test failed"
    echo "See integration test section for test runner logs, and uploaded artifacts for server logs."
  fi
}
