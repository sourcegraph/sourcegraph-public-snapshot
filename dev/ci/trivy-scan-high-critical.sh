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

trivy_scan() {
  local templateFile="$1"
  local outputFile="$2"
  local target="$3"

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
    "@${templateFile}"

    # dump the HTML output to a file named "ANNOTATION_FILE"
    "--output"
    "${outputFile}"

    # scan the docker image named "target"
    "${target}"
  )

  trivy image "${TRIVY_ARGS[@]}"
}

ARTIFACT_FILE="${IMAGE}-security-report.html"
if ! trivy_scan "./dev/ci/trivy-artifact-html.tpl" "${OUTPUT}/${ARTIFACT_FILE}" "${IMAGE}"; then

  pushd "${OUTPUT}"
  buildkite-agent artifact upload "${ARTIFACT_FILE}"

  cat <<EOF | buildkite-agent annotate --style warning --context "Docker Image security scan"
      The \`${IMAGE}\` Docker image has \`HIGH/CRITICAL\` severity CVE(s): <a href="artifact://${ARTIFACT_FILE}">security scan report</a>
EOF
  popd

  exit 1
fi
