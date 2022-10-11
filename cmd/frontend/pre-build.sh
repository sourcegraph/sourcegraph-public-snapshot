#!/usr/bin/env bash
set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

# Build the webapp typescript code.
echo "--- yarn"
# mutex is necessary since CI runs various yarn installs in parallel
if [[ -z "${CI}" ]]; then
  yarn
else
  ./dev/ci/yarn-install-with-retry.sh
fi

echo "--- yarn run build-web"
NODE_ENV=production DISABLE_TYPECHECKING=true yarn run build-web
