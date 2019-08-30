#!/usr/bin/env bash

set -e

echo 'ENTERPRISE='$ENTERPRISE

cd $1
echo "--- yarn"
yarn --frozen-lockfile --network-timeout 60000

echo "--- build"
yarn -s run build --color
