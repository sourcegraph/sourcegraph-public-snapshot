#!/usr/bin/env bash

set -e

echo "ENTERPRISE=$ENTERPRISE"
echo "NODE_ENV=$NODE_ENV"
echo "# Note: NODE_ENV only used for build command"

echo "--- pnpm install in root"
# mutex is necessary since CI runs various pnpm installs in parallel
NODE_ENV='' pnpm install

cd "$1"
echo "--- browserslist"
NODE_ENV='' pnpm --silent run browserslist

echo "--- build"
pnpm --silent run build --color

# Only run bundlesize if intended and if there is valid a script provided in the relevant package.json
if [ "$CHECK_BUNDLESIZE" ] && jq -e '.scripts.bundlesize' package.json >/dev/null; then
  echo "--- bundlesize"
  pnpm --silent run bundlesize --enable-github-checks
fi
