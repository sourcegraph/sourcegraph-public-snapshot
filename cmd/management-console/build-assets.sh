#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -ex

# for node_modules/@sourcegraph/tsconfig/tsconfig.json
pushd ../..
yarn --frozen-lockfile
popd

pushd web/
yarn run build
popd
