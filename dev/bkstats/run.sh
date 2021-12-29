#! /usr/bin/env bash

function usage {
  cat <<EOF
Usage: run.sh

Post a message on Slack reporting how much time the main branch was red yesterday.

Requires:
- \$BUILDKITE_TOKEN
- \$WEBHOOK_URL
EOF
}

if [ "$1" == "-h" ]; then
  usage
  exit 1
fi

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

pushd dev/bkstats

echo "--- Posting report on Slack"
go run main.go -buildkite.token="$BUILDKITE_TOKEN" -slack.webhook="$WEBHOOK_URL"

popd
