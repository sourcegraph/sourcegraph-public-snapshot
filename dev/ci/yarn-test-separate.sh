#!/usr/bin/env bash

set -e

echo "--- yarn in root"
# mutex is necessary since CI runs various pnpm installs in parallel
pnpm install

cd "$1"
echo "--- yarn"
# mutex is necessary since CI runs various pnpm installs in parallel
pnpm install

echo "--- test"

# Limit the number of workers to prevent the default of 1 worker per core from
# causing OOM on the buildkite nodes that have 96 CPUs. 4 matches the CPU limits
# in infrastructure/kubernetes/ci/buildkite/buildkite-agent/buildkite-agent.Deployment.yaml
yarn --silent run test --maxWorkers 4
