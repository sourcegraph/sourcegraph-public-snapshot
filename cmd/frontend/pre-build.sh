#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Build the webapp typescript code.
pushd cmd/frontend/internal/app/web_modules
yarn install
yarn run fmt-check
yarn run lint
yarn run build
popd

go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
