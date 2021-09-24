#!/usr/bin/env bash

set -e

echo "ENTERPRISE=$ENTERPRISE"
echo "NODE_ENV=$NODE_ENV"
echo "# Note: NODE_ENV only used for build command"

echo "--- Yarn install in root"
# mutex is necessary since CI runs various yarn installs in parallel
NODE_ENV='' yarn --mutex network --frozen-lockfile --network-timeout 60000

cd "$1"
echo "--- browserslist"
NODE_ENV='' yarn -s run browserslist

echo "--- build"
yarn -s run build --color
