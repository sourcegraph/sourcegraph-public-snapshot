#!/usr/bin/env bash

function usage {
  cat <<EOF
Usage: upload-build-logs.sh

Upload a buildkite build result in Loki.

Requires:
- \$BUILDKITE_COMMIT
- \$LOKI_URL
EOF
}

if [ "$1" == "-h" ]; then
  usage
  exit 1
fi

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

echo "--- :go: Building sg"
(
  set -x
  pushd dev/sg
  go build -o ../../ci_sg -ldflags "-X main.BuildCommit=$BUILDKITE_COMMIT" -mod=mod .
  popd
)

echo "--- :arrow_up: Uploading logs (if build failed)"
./ci_sg ci logs --out=$LOKI_URL --state="failed"
