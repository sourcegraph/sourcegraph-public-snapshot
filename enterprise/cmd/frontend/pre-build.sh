#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

pushd ..
echo "--- yarn install"
yarn --frozen-lockfile --network-timeout 60000
echo "--- build browser extension and code host integrations"
(pushd browser && TARGETS=phabricator yarn build && popd)
echo "--- build web app"
(pushd web && NODE_ENV=production yarn -s run build --color && popd)
popd

echo "--- generate"
dev/generate.sh
