#!/usr/bin/env bash

set -eu

# Calls the enterprise/dev/app/release.sh script with the right env vars.

export GITHUB_TOKEN=$(gcloud secrets versions access latest --secret=BUILDKITE_GITHUBDOTCOM_TOKEN --quiet --project=sourcegraph-ci)

TMPFILE=$(mktemp)
gcloud secrets versions access latest --secret=BUILDKITE_GCLOUD_SERVICE_ACCOUNT --quiet --project=sourcegraph-ci > $TMPFILE
export GCLOUD_APP_CREDENTIALS_FILE=$TMPFILE
cleanup() {
  rm "$TMPFILE"
}
trap cleanup EXIT

ROOTDIR="$(realpath $(dirname "${BASH_SOURCE[0]}")/../../../..)"
exec "$ROOTDIR"/enterprise/dev/app/release.sh
