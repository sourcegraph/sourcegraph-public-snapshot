#!/usr/bin/env bash

set -e

echo "--- Find a commit to compare the bundle size against"
MERGE_BASE=$(git merge-base HEAD origin/main)
export MERGE_BASE

# We may not have a stats.json file for the merge base commit as these are only
# created for commits that touch frontend files. Instead, we scan for 20 commits
# before the merge base and use the latest stats.json file we find.
REVISIONS=()
while IFS='' read -r line; do REVISIONS+=("$line"); done < <(git --no-pager log "${MERGE_BASE}" --pretty=format:"%H" -n 20)
for REVISION in "${REVISIONS[@]}"; do
  gsutil -q cp -r "gs://sourcegraph_buildkite_cache/sourcegraph/sourcegraph/bundle_size_cache-$REVISION.tar.gz" "./ui/assets/bundle_size_cache-$REVISION.tar.gz" || echo "Cached archive for $REVISION not found, skipping."
  if [[ -f "./ui/assets/bundle_size_cache-$REVISION.tar.gz" ]]; then
    echo "Found cached archive for $REVISION"
    export COMPARE_REV=$REVISION
    break
  fi
done
tar -xf "ui/assets/bundle_size_cache-${COMPARE_REV}.tar.gz" ui/assets || echo "Could not extract archive for $COMPARE_REV"

echo "--- Report bundle diff"

ls -la ./ui/assets/

COMMIT_FILE="./ui/assets/stats-${COMMIT}.json"
COMPARE_FILE="./ui/assets/stats-${COMPARE_REV}.json"
if [[ -f $COMMIT_FILE ]] && [[ -f $COMPARE_FILE ]]; then
  ./node_modules/.bin/statoscope generate \
    -i "${COMMIT_FILE}" \
    -r "${COMPARE_FILE}" \
    -t ./ui/assets/compare-report.html

  echo "gs://sourcegraph_reports/statoscope-reports/${BRANCH}"
  gsutil cp ./ui/assets/compare-report.html "gs://sourcegraph_reports/statoscope-reports/${BRANCH}/compare-report.html"

  echo "${COMMIT_FILE}" "${COMPARE_FILE}"
  pnpm --filter @sourcegraph/web run report-bundle-diff \
    "${COMMIT_FILE}" \
    "${COMPARE_FILE}"
else
  echo 'No stats file found, skipping.'
  echo "$COMMIT_FILE"
  echo "$COMPARE_FILE"
fi
