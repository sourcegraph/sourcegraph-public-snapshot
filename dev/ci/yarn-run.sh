#!/usr/bin/env bash

set -e

echo "--- yarn"
./dev/ci/yarn-install-with-retry.sh
./dev/ci/yarn-install-with-retry.sh --cwd dev/release

echo "--- generate"
yarn gulp generate

for cmd in "$@"; do
  echo "--- $cmd"
  yarn run "$cmd"
done
