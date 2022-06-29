#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

created_from=$(date -d "7 days ago" '+%Y-%m-%d')
created_to=$(date -d "2 days ago" '+%Y-%m-%d')

echo "--- Running 'buildchecker history' from $created_from to $created_to"
go run ./dev/buildchecker/ \
  -buildkite.token="$BUILDKITE_TOKEN" \
  -created.from="$created_from" \
  -created.to="$created_to" \
  -slack.token="$SLACK_TOKEN" \
  -slack.report-webhook="$SLACK_REPORT_WEBHOOK" \
  history
