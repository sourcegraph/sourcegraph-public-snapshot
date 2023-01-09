#!/usr/bin/env bash

set -e

echo "--- pnpm install in root"
./dev/ci/pnpm-install-with-retry.sh

# Save the absolute path to the script before changing directories.
abs_pnpm="$(pwd)/dev/ci/pnpm-install-with-retry.sh"

cd "$1"
echo "--- pnpm install"
"$abs_pnpm"

echo "--- test"

# Limit the number of workers to prevent the default of 1 worker per core from
# causing OOM on the buildkite nodes that have 96 CPUs. 4 matches the CPU limits
# in infrastructure/kubernetes/ci/buildkite/buildkite-agent/buildkite-agent.Deployment.yaml
pnpm -s run test --maxWorkers 4
