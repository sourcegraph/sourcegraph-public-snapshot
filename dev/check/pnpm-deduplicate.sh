#!/usr/bin/env bash
set -eu -o pipefail

echo "--- check pnpm-lock.yaml for duplicates"

# Prevent duplicates in yarn.lock/node_modules that lead to errors and bloated bundle sizes

./dev/ci/pnpm-install-with-retry.sh

echo "Checking for duplicate dependencies in pnpm-lock.yaml"