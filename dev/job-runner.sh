#!/usr/bin/env bash

# Description: Runs the ephemeral job-runner container.

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

function finish() {
  echo 'trapped CTRL-C: stopping docker job-runner container'
  docker rm -f job-runner
}
[ -n "${WATCH_TRIGGER-}" ] || trap finish EXIT

# Stop the previously running container, if present.
docker inspect job-runner >/dev/null 2>&1 && docker rm -f job-runner

# Build the image
rm -rf .bin/job-runner-tmp && mkdir -p .bin/job-runner-tmp && OUTPUT=.bin/job-runner-tmp IMAGE="sourcegraph/job-runner:dev" ./cmd/job-runner/build.sh

# Run the container
docker run \
  --name=job-runner \
  --cpus=1 \
  --memory=4g \
  --restart=always \
  -p 0.0.0.0:3190:3190 \
  sourcegraph/job-runner:dev &

[ -n "${WATCH_TRIGGER-}" ] || sleep 99999999
