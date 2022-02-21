#!/usr/bin/env bash
set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

# Build the webapp typescript code.
echo "--- yarn"
# mutex is necessary since CI runs various yarn installs in parallel
yarn install

echo "--- yarn run build-web"
NODE_ENV=production DISABLE_TYPECHECKING=true yarn run build-web
