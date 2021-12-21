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

if hash go 2>/dev/null; then
  echo "--- :go: Building sg"
  (
    set -x
    pushd dev/sg
    go build -o ../../ci_sg -ldflags "-X main.BuildCommit=$BUILDKITE_COMMIT" -mod=mod .
    popd
  )
else
  # The following code is a series a workaround to fix issues with the buildkite agents.
  # TODO(@jhchabran) clean this once we've fixed the baremetal buildkite agents.
  echo "--- :arrow_down: Downloading sg"
  (
    set -x
    _location_header="$(curl --silent -I "https://github.com/sourcegraph/sg/releases/latest" | grep "location:" | tr -d '\r')"
    _base_url="$(echo "${_location_header}" | sed s/location:\ // | sed s/tag/download/ | tr -d "[:blank:]")"
    wget "$_base_url"/sg_linux_amd64 -O ci_sg
    chmod +x ci_sg
  )

  BUILDKITE_API_TOKEN=$(gcloud secrets versions access latest --secret="BUILDKITE_API_TOKEN" --project="sourcegraph-ci" --quiet)
  BUILD_LOGS_LOKI_URL=$(gcloud secrets versions access latest --secret="BUILD_LOGS_LOKI_URL" --project="sourcegraph-ci" --quiet)
  export BUILDKITE_API_TOKEN
  export BUILD_LOGS_LOKI_URL
fi

echo "--- :file_cabinet: Uploading logs"
# Because we are running this script in the buildkite post-exit hook, the state of the job is still "running".
# Passing --state="" just overrides the default. It's not set to any specific state because this script caller
# is responsible of making sure the job has failed.
./ci_sg --pristine-limits ci logs --out="$BUILD_LOGS_LOKI_URL" --state="" --overwrite-state="failed" --build="$BUILDKITE_BUILD_NUMBER" --job="$BUILDKITE_JOB_ID"
