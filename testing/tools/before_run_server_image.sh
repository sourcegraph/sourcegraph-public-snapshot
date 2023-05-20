#!/bin/bash

echo "--- Ensuring clean slate before running server"

if [ "$BUILDKITE" != true ]; then
  echo "⚠️ This script is NOT running on a Buildkite agent."
  echo "👉 It will therefore NOT clean anything."
  exit 0
fi

echo "--- Ensuring clean slate before running server container"

running=$(docker ps -aq | wc -l)
if [[ "$running" -gt 0 ]]; then
  echo "⚠️ Found $running running containers, deleting them."
  # shellcheck disable=SC2046
  docker rm -f $(docker ps -aq)
else
  echo "Found 0 running containers."
fi

images=$(docker images -q | wc -l)
if [[ "$images" -gt 0 ]]; then
  echo "⚠️ Found $images images, deleting them."
  # shellcheck disable=SC2046
  docker rmi -f $(docker images -q)
else
  echo "Found 0 images."
fi

echo "Removing existing volumes, if any"
docker volume prune -f

echo "--- "
