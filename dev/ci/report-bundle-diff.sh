#!/usr/bin/env bash

set -e

echo "--- Report bundle diff"

ls -la ./ui/assets/

commitFile="./ui/assets/stats-${COMMIT}.json"
mergeBaseFile="./ui/assets/stats-${MERGE_BASE}.json"
if [[ -f $commitFile ]] && [[ -f $mergeBaseFile ]]; then
  ./node_modules/.bin/statoscope generate \
    -i "${commitFile}" \
    -r "${mergeBaseFile}" \
    -t ./ui/assets/compare-report.html

  gsutil cp ./ui/assets/compare-report.html "gs://sourcegraph_reports/statoscope-reports/${BRANCH}"

  yarn workspace @sourcegraph/web run report-bundle-diff \
    "${commitFile}" \
    "${mergeBaseFile}"
else
  echo 'No stats file found, skipping.'
fi
