#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../../..
set -euo pipefail

export GITHUB_TOKEN="${GH_TOKEN}"

# do not move this "set -x" above the GITHUB_TOKEN
# env var alias above, we don't want this to leak
# inside of CI's logs
set -x

# This is the special exit code that we tell trivy to use
# if finds a vulnerability

trivy_scan() {
  local templateFile="$1"
  local outputFile="$2"
  local target="$3"

  TRIVY_ARGS=(
    # fail the step if there is a vulnerability
    "--exit-code"
    "${VULNERABILITY_EXIT_CODE}"

    # ignore issues that we can't fix
    "--ignore-unfixed"

    # we'll only take action on higher CVEs
    "--severity"
    "HIGH,CRITICAL"

    # tell trivy to dump its output to an HTML file
    "--format"
    "template"

    # use the custom "trivy-html" template that we have in this folder
    "--template"
    "@${templateFile}"

    # dump the HTML output to a file named "outputFile"
    "--output"
    "${outputFile}"

    # workaround for scans failing due to exceeding timeout deadline
    # (e.g. https://buildkite.com/sourcegraph/sourcegraph/builds/192926#0185a568-ee3e-494a-abbf-7b8f9c3f226f/118-173)
    "--timeout"
    "15m"

    # scan the docker image named "target"
    "${target}"
  )

  trivy image "${TRIVY_ARGS[@]}"
}

create_annotation() {
  local path="$1"
  local imageName="$2"

  local file
  file="$(basename "${path}")"

  cat <<EOF >./annotations/trivy-scan-high-critical.md
- **${imageName}** high/critical CVE(s): [${file}](artifact://${file})
EOF

  echo "High or critical severity CVEs were discovered in ${IMAGE}. Please read the buildkite annotation for more info."
}

ARTIFACT_FILE="$(pwd)/${IMAGE}-security-report.html"
trivy_scan "./dev/ci/trivy/trivy-artifact-html.tpl" "${ARTIFACT_FILE}" "${IMAGE}" || exitCode="$?"
case "${exitCode:-"0"}" in
  0)
    # no vulnerabilities were found
    exit 0
    ;;
  "${VULNERABILITY_EXIT_CODE}")
    # we found vulnerabilities - upload the annotation
    create_annotation "${ARTIFACT_FILE}" "${IMAGE}"
    echo "<br />Trivy found issues. More information [here](https://handbook.sourcegraph.com/departments/product-engineering/engineering/cloud/security/trivy/#for-sourcegraph-engineers)." >> ./annotations/trivy-scan-high-critical.md
    exit "${VULNERABILITY_EXIT_CODE}"
    ;;
  *)
    # some other kind of error occurred
    exit $exitCode
    ;;
esac
