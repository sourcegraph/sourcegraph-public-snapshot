#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

dev/generate.sh

pushd ..
yarn --frozen-lockfile --network-timeout 60000
(cd web && yarn -s run build --color)
popd
