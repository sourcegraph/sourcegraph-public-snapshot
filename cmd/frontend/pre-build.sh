#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Build the webapp typescript code.
[[ -z "${CI}" ]] && npm install || npm ci
pushd web
[[ -z "${CI}" ]] && npm install || npm ci
NODE_ENV=production DISABLE_TYPECHECKING=true npm run build
popd

go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
