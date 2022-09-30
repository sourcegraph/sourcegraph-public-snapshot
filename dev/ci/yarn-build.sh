#!/usr/bin/env bash

set -e

echo "ENTERPRISE=$ENTERPRISE"
echo "NODE_ENV=$NODE_ENV"
echo "# Note: NODE_ENV only used for build command"

echo "--- Yarn install in root"
NODE_ENV='' ./dev/ci/yarn-install-with-retry.sh

cd "$1"
echo "--- browserslist"
NODE_ENV='' yarn run browserslist

echo "--- build"
yarn run build --color

# Only run bundlesize if intended and if there is valid a script provided in the relevant package.json
if [ "$CHECK_BUNDLESIZE" ] && jq -e '.scripts.bundlesize' package.json >/dev/null; then
  echo "--- bundlesize"
  yarn run bundlesize --enable-github-checks
  echo "--- bundlesize:web:upload"
  HONEYCOMB_API_KEY="$CI_HONEYCOMB_CLIENT_ENV_API_KEY" yarn workspace @sourcegraph/observability-server run bundlesize:web:upload

  if [[ "$BRANCH" != "main" ]]; then
    echo "--- generate statoscope comparison report"
    pushd "../.." >/dev/null

    ls -la ./ui/assets/

    commitFile="./ui/assets/stats-${COMMIT}.json"
    mergeBaseFile="./ui/assets/stats-${MERGE_BASE}.json"
    if [[ -f $commitFile ]] && [[ -f $mergeBaseFile ]]; then
      ./node_modules/.bin/statoscope generate \
        -i "${commitFile}" \
        -r "${mergeBaseFile}" \
        -t ./ui/assets/compare-report.html

      yarn workspace @sourcegraph/web run report-bundle-diff \
        "${commitFile}" \
        "${mergeBaseFile}"
    else
      echo 'No stats file found, skipping.'
    fi

    popd >/dev/null || exit
  fi
fi
