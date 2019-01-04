#!/usr/bin/env bash

set -e

echo "--- yarn in root"
yarn --frozen-lockfile --network-timeout 60000

cd $1
echo "--- test"
yarn -s run test

