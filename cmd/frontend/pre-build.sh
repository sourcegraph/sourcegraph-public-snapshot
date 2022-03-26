#!/usr/bin/env bash
set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

# Build the webapp typescript code.
echo "--- yarn"
# mutex is necessary since CI runs various pnpm installs in parallel
pnpm install

echo "--- pnpm runbuild-web"
NODE_ENV=production DISABLE_TYPECHECKING=true pnpm run build-web
