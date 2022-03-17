#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

NAME="deployment-notifier"

OUTPUT=$(mktemp -d -t dn_build_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

GOOS=linux GOARCH=amd64 go build -o "$OUTPUT/$NAME"
gzip "$OUTPUT/$NAME"

gsutil cp "$OUTPUT/$NAME.gz" gs://sourcegraph_buildkite_cache/
