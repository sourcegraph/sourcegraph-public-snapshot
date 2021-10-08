#!/usr/bin/env bash

set -euo pipefail

OUTPUT=$(mktemp -d -t trivy_XXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

export GITHUB_TOKEN="${GH_TOKEN}"

# do not move this "set -x" above the GITHUB_TOKEN
# env var alias above, we don't want this to leak
# inside of CI's logs
set -x

ANNOTATION_FILE="${OUTPUT}/annotation.md"

if ! trivy image "$@" -o "${ANNOTATION_FILE}"; then
  buildkite-agent annotate --style warning --context "${APP} Docker Image security scan" <"${ANNOTATION_FILE}"
  exit 1
fi
