#!/usr/bin/env bash

set -e

echo "--- yarn in root"
./dev/ci/yarn-install-with-retry.sh

# Save the absolute path to the script before changing directories.
abs_yarn="$(pwd)/dev/ci/yarn-install-with-retry.sh"

cd "$1"
echo "--- yarn"
"$abs_yarn"

echo "--- test"

# Limit the number of workers to prevent the default of 1 worker per core from
# causing OOM on the buildkite nodes that have 96 CPUs. 4 matches the CPU limits
# in infrastructure/kubernetes/ci/buildkite/buildkite-agent/buildkite-agent.Deployment.yaml
yarn -s run test --maxWorkers 4
