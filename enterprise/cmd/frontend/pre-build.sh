#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

pushd ..
echo "--- yarn install"
yarn --frozen-lockfile --network-timeout 60000
echo "--- browser: yarn run build (phabricator integration assets)"
(pushd browser && TARGETS=phabricator yarn build && popd)
echo "--- web: yarn run build"
(pushd web && NODE_ENV=production yarn -s run build --color && popd)
popd

echo "--- generate"
dev/generate.sh
