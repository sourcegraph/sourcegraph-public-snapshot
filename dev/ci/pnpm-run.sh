#!/usr/bin/env bash

set -e

echo "--- pnpm install"
./dev/ci/pnpm-install-with-retry.sh

echo "--- generate"
pnpm run generate

for cmd in "$@"; do
  echo "--- $cmd"
  pnpm run "$cmd"
done
