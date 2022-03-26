#!/usr/bin/env bash

set -e

echo "--- yarn"
# mutex is necessary since CI runs various pnpm installs in parallel
pnpm install
# yarn --mutex network --cwd dev/release --immutable --network-timeout 60000

echo "--- generate"
pnpm gulp generate

for cmd in "$@"; do
  echo "--- $cmd"
  pnpm --silent run "$cmd"
done
