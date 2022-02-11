#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

echo "--- Running 'pr-auditor'"
go run ./dev/pr-auditor/ \
  -github.payload-file="$GITHUB_EVENT_PATH" \
  -github.token="$GITHUB_TOKEN"
