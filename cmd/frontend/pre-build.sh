#!/usr/bin/env bash
set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

# Build the webapp typescript code.
echo "--- pnpm"
# mutex is necessary since CI runs various pnpm installs in parallel
if [[ -z "${CI}" ]]; then
  pnpm install
else
  pnpm install  --frozen-lockfile
fi

echo "--- pnpm build-web"
NODE_ENV=production DISABLE_TYPECHECKING=true pnpm build-web
