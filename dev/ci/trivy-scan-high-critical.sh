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

ANNOTATION_FILE="${OUTPUT}/annotation.html"

TRIVY_ARGS=(
  # fail the step if there is a vulnerability
  "--exit-code"
  "1"

  # ignore issues that we can't fix
  "--ignore-unfixed"

  # we'll only take action on higher CVEs
  "--severity"
  "HIGH,CRITICAL"

  # tell trivy to dump its output to an HTML file
  "--format"
  "template"

  # use the custom "trivy-html" that we have in this folder
  "--template"
  "@./dev/ci/trivy-html.tpl"

  # dump the HTML output to a file named "ANNOTATION_FILE"
  "--output"
  "${ANNOTATION_FILE}"

  # scan the docker image named "IMAGE"
  "${IMAGE}"
)

if ! trivy image "${TRIVY_ARGS[@]}"; then
  buildkite-agent annotate --style warning --context "${APP} Docker Image security scan" <"${ANNOTATION_FILE}"
  exit 1
fi
