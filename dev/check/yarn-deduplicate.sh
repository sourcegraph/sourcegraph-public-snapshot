#!/usr/bin/env bash
set -eu -o pipefail

echo "--- check yarn.lock for duplicates"

# Prevent duplicates in yarn.lock/node_modules that lead to errors and bloated bundle sizes

./dev/ci/yarn-install-with-retry.sh --ignore-scripts

echo "Checking for duplicate dependencies in yarn.lock"
yarn run -s yarn-deduplicate --fail --list --strategy fewer ./yarn.lock || {
  echo 'yarn.lock contains duplicate dependencies. Please run "yarn deduplicate" and commit the result.'
  echo "^^^ +++"
  exit 1
}
