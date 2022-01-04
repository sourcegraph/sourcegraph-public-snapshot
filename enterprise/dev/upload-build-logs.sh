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

source /root/.profile

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

echo "--- :go: Building sg"
(
  set -x
  pushd dev/sg
  go build -o ../../ci_sg -ldflags "-X main.BuildCommit=$BUILDKITE_COMMIT" -mod=mod .
  popd
)

echo "--- :file_cabinet: Uploading logs"

# Because we are running this script in the buildkite post-exit hook, the state of the job is still "running".
# Passing --state="" just overrides the default. It's not set to any specific state because this script caller
# is responsible of making sure the job has failed.
./dev/ci/sentry-capture.sh ./ci_sg ci logs --out="$BUILD_LOGS_LOKI_URL" --state="" --overwrite-state="failed" --build="$BUILDKITE_BUILD_NUMBER" --job="$BUILDKITE_JOB_ID"
