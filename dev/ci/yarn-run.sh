#!/usr/bin/env bash

set -e

echo "--- yarn"
# mutex is necessary since CI runs various yarn installs in parallel
pnpm --frozen-lockfile
pnpm --C dev/release --frozen-lockfile

echo "--- generate"
pnpm gulp generate

for cmd in "$@"; do
  echo "--- $cmd"
  pnpm -s run "$cmd"
done
