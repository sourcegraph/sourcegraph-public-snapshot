#!/usr/bin/env bash

set -e

echo "--- yarn in root"
yarn --frozen-lockfile --network-timeout 60000

cd $1
echo "--- cover"
yarn -s run cover

echo "--- report"
yarn -s run nyc report -r json --report-dir coverage


