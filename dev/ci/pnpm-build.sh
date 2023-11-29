#!/usr/bin/env bash

set -e

echo "NODE_ENV=$NODE_ENV"
echo "# Note: NODE_ENV only used for build command"

echo "--- Pnpm install in root"
NODE_ENV='' ./dev/ci/pnpm-install-with-retry.sh

cd "$1"

echo "--- build"
pnpm run build --color

# Only run bundlesize if intended and if there is valid a script provided in the relevant package.json
if [ "$CHECK_BUNDLESIZE" ] && jq -e '.scripts.bundlesize' package.json >/dev/null; then
  echo "--- bundlesize:web:upload"
  HONEYCOMB_API_KEY="$CI_HONEYCOMB_CLIENT_ENV_API_KEY" pnpm --filter @sourcegraph/observability-server run bundlesize:web:upload
fi
