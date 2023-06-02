#!/usr/bin/env bash
set -eu -o pipefail

echo "--- check pnpm-lock.yaml for duplicates"

# Prevent duplicates in pnpm-lock.yaml/node_modules that lead to errors and bloated bundle sizes

./dev/ci/pnpm-install-with-retry.sh
echo "Checking for duplicate dependencies in pnpm-lock.yaml"

pnpm dedupe --check || {
  echo 'pnpm-lock.yaml contains duplicate dependencies. Please run "pnpm dedupe" and commit the result.'
  echo "^^^ +++"
  exit 1
}
