#!/usr/bin/env bash

set -e

echo "--- yarn"
# mutex is necessary since CI runs various yarn installs in parallel
yarn install
# yarn --mutex network --cwd dev/release --immutable --network-timeout 60000

echo "--- generate"
yarn gulp generate

for cmd in "$@"; do
  echo "--- $cmd"
  yarn --silent run "$cmd"
done
