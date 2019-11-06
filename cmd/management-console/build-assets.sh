#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -euxo pipefail

# for node_modules/@sourcegraph/tsconfig/tsconfig.json
pushd ../..

echo "--- go build"
yarn --mutex network install

popd

pushd web/

echo "--- yarn install web"
yarn --mutex network install

echo "-- yarn build web"
yarn run build

popd
