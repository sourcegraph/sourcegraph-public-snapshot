#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -ex

# for node_modules/@sourcegraph/tsconfig/tsconfig.json
pushd ../..
yarn install
popd

pushd web/
yarn install
yarn run build
popd
