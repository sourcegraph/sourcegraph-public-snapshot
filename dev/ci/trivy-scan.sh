#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

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

# download html template
# HTML_TEMPLATE_FILE="${OUTPUT}/html.tpl"

# TRIVY_VERSION="${TRIVY_VERSION:-0.20.0}"
# curl "https://raw.githubusercontent.com/aquasecurity/trivy/v${TRIVY_VERSION}/contrib/html.tpl" >"${HTML_TEMPLATE_FILE}"

ANNOTATION_FILE="${OUTPUT}/annotation.html"
# ANNOTATION_MARKDOWN="${OUTPUT}/annotation.md"

if ! trivy image --format template --template "@./dev/ci/trivy-html.tpl" -o "${ANNOTATION_FILE}" "$@"; then
  # pandoc "${ANNOTATION_FILE}" --to gfm -o "${ANNOTATION_MARKDOWN}"
  buildkite-agent annotate --style warning --context "${APP} Docker Image security scan" <"${ANNOTATION_FILE}"
  exit 1
fi
