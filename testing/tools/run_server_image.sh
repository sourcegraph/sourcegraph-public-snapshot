#!/bin/bash

set -eu

# Data folder for the server image
DATA="tmp_run_server_image"
mkdir "$DATA"
DATA="$(pwd)/$DATA"

# Path to the tarball of the server image
image_tarball="$1"
image_name="$2"

# shellcheck disable=SC2124
args="${@:3}"

# Server configuration
PORT=${PORT:-"7080"}
URL="http://localhost:$PORT"
SOURCEGRAPH_LICENSE_GENERATION_KEY=${SOURCEGRAPH_LICENSE_GENERATION_KEY:-""}
DB_STARTUP_TIMEOUT="10s"
ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="true"

# Feature flags
SG_FEATURE_FLAG_GRPC=${SG_FEATURE_FLAG_GRPC:-"false"}

echo "--- Checking for existing server instance at $URL"

if curl --output /dev/null --silent --head --fail "$URL"; then
  echo "❌ Can't run a new server instance on $URL because another instance is already running."
  exit 1
fi

echo "--- Loading server image"
echo "Loading $image_tarball in Docker"
docker load --input "$image_tarball"

echo "-- Starting $image_name"
echo "Listening at: $URL"
echo "Data and config volume bounds: $DATA"
echo "Database startup timeout: $DB_STARTUP_TIMEOUT"

echo "Allow single docker image code insights: $ALLOW_SINGLE_DOCKER_CODE_INSIGHTS"
echo "GRPC Feature flag: $SG_FEATURE_FLAG_GRPC"

docker run $args \
  --publish "$PORT":7080 \
  -e ALLOW_SINGLE_DOCKER_CODE_INSIGHTS="$ALLOW_SINGLE_DOCKER_CODE_INSIGHTS" \
  -e SOURCEGRAPH_LICENSE_GENERATION_KEY="$SOURCEGRAPH_LICENSE_GENERATION_KEY" \
  -e SG_FEATURE_FLAG_GRPC="$SG_FEATURE_FLAG_GRPC" \
  -e DB_STARTUP_TIMEOUT="$DB_STARTUP_TIMEOUT" \
  --volume "$DATA/config:/etc/sourcegraph" \
  --volume "$DATA/data:/var/opt/sourcegraph" \
  "$image_name"

echo "-- Listening at $URL"
