#!/usr/bin/env bash

function usage {
  cat <<EOF
Usage: upload-build-logs.sh

Upload a buildkite build result in Loki.

Requires:
- \$BUILDKITE_BUILD_NUMBER
- \$BUILDKITE_JOB_ID
- \$BUILD_LOGS_LOKI_URL
EOF
}

if [ "$1" == "-h" ]; then
  usage
  exit 1
fi

# shellcheck disable=SC1091
source /root/.profile

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"/../../../..

echo "~~~ :go: Building sg"
(
  set -x
  pushd dev/sg
  go build -o ../../sg -mod=mod .
  popd
)

echo "~~~ :file_cabinet: Uploading logs"

# Because we are running this script in the buildkite post-exit hook, the state of the job is still "running".
# Passing --state="" just overrides the default. It's not set to any specific state because this script caller
# is responsible of making sure the job has failed.
export SG_DISABLE_OUTPUT_DETECTION=true
./dev/ci/scripts/sentry-capture.sh ./sg ci logs --out="$BUILD_LOGS_LOKI_URL" --state="" --overwrite-state="failed" --build="$BUILDKITE_BUILD_NUMBER" --job="$BUILDKITE_JOB_ID"
local_exit_code=$?
if [[ $local_exit_code -ne 0 ]]; then
  echo -e "\033[33m┌────────────────────────────────────────────────────────────────────┐\033[0m"
  echo -e "\033[33m│ The failure in this hook does not impact the outcome of this build │\033[0m"
  echo -e "\033[33m└────────────────────────────────────────────────────────────────────┘\033[0m"
fi
exit $local_exit_code
