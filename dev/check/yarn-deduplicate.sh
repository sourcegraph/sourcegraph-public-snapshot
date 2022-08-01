#!/usr/bin/env bash
set -eu -o pipefail

echo "--- check yarn.lock for duplicates"

# Prevent duplicates in yarn.lock/node_modules that lead to errors and bloated bundle sizes

./dev/ci/yarn-install-with-retry.sh

echo "Checking for duplicate dependencies in yarn.lock"
yarn dedupe --check || {
  echo 'yarn.lock contains duplicate dependencies. Please run "yarn dedupe" and commit the result.'
  echo "^^^ +++"
  exit 1
}
