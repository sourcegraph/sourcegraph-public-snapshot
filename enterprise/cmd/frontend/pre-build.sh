#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

pushd ..
yarn --frozen-lockfile
NODE_ENV=production yarn run build --color
popd

dev/generate.sh
