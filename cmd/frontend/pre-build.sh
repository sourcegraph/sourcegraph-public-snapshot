#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Build the webapp typescript code.
echo "--- yarn"
# mutex is necessary since CI runs various yarn installs in parallel
[[ -z "${CI}" ]] && yarn --mutex network || yarn --mutex network --frozen-lockfile --network-timeout 60000

pushd web
echo "--- yarn run build"
NODE_ENV=production DISABLE_TYPECHECKING=true yarn run build
popd

echo "--- go generate"
go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
