#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

pushd ..
echo "--- yarn root"
yarn --frozen-lockfile --network-timeout 60000
echo "--- yarn browser"
(pushd browser && TARGETS=phabricator yarn build && popd)
echo "--- yarn web"
(pushd web && NODE_ENV=production yarn -s run build --color && popd)
popd

echo "--- generate"
dev/generate.sh
