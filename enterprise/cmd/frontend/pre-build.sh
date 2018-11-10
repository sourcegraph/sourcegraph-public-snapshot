#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

pushd ..
yarn --frozen-lockfile --network-timeout 60000
yarn workspace sourcegraph run build
yarn workspace @sourcegraph/extensions-client-common run build
yarn workspace webapp run build --color
popd

dev/generate.sh
