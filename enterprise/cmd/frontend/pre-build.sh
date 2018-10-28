#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Generate ../dist in the OSS dependency
pushd ..
yarn --frozen-lockfile
NODE_ENV=production yarn run dist
popd

# Build the enterprise webapp typescript code.
yarn --frozen-lockfile
NODE_ENV=production yarn run build --color

dev/generate.sh
