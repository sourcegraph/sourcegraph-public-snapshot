#!/usr/bin/env bash

# Use the shared volume in between dind and the agent to host the data, so we can delete it afterward.
export DATA="/mnt/tmp/sourcegraph-data"

echo Y | ./dev/run-server-image.sh -d --name sourcegraph

SOURCEGRAPH_BASE_URL="http://localhost:7080"
echo "--- Waiting for $SOURCEGRAPH_BASE_URL to be up"
set +e
timeout 120s bash -c "until curl --output /dev/null --silent --head --fail $SOURCEGRAPH_BASE_URL; do
  echo Waiting 5s for $SOURCEGRAPH_BASE_URL...
  sleep 5
done"

echo "--- Up, running tests..."
