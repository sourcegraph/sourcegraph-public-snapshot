#!/usr/bin/env bash

set -eu

# Calls the dev/app/release.sh script with the right env vars.

GITHUB_TOKEN=$(gcloud secrets versions access latest --secret=BUILDKITE_GITHUBDOTCOM_TOKEN --quiet --project=sourcegraph-ci)
export GITHUB_TOKEN

# TODO(sqs): Make this non-optional (by removing ` || echo -n`) when https://github.com/sourcegraph/infrastructure/pull/4481 is merged.
SLACK_APP_RELEASE_WEBHOOK=$(gcloud secrets versions access latest --secret=SLACK_APP_RELEASE_WEBHOOK --quiet --project=sourcegraph-ci || echo -n)
export SLACK_APP_RELEASE_WEBHOOK

TMPFILE=$(mktemp)
gcloud secrets versions access latest --secret=BUILDKITE_GCLOUD_SERVICE_ACCOUNT --quiet --project=sourcegraph-ci > "$TMPFILE"
export GCLOUD_APP_CREDENTIALS_FILE=$TMPFILE
cleanup() {
  rm "$TMPFILE"
}
trap cleanup EXIT

ROOTDIR="$(realpath "$(dirname "${BASH_SOURCE[0]}")"/../../../..)"
exec "$ROOTDIR"/dev/app/release.sh
