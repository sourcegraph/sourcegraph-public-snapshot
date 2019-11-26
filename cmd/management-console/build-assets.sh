#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -euxo pipefail

# for node_modules/@sourcegraph/tsconfig/tsconfig.json
pushd ../..
# mutex is necessary since frontend and the management-console can
# run concurrent "yarn" installs
yarn --mutex network install
popd

pushd web/
# mutex is necessary since frontend and the management-console can
# run concurrent "yarn" installs
yarn --mutex network install
yarn run build
popd
