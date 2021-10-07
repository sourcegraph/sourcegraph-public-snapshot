#!/usr/bin/env bash

set -euo pipefail

export GITHUB_TOKEN="${GH_TOKEN}"

# do not move this "set -x" above the GITHUB_TOKEN
# env var alias above, we don't want this to leak
# inside of CI's logs
set -x

OUTPUT=$(mktemp -d -t trivy_XXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

pushd "${OUTPUT}"

SARIF_REPORT_FILE="report.sarif"

# download sarif template
SARIF_TEMPLATE_FILE="sarif.tpl"

TRIVY_VERSION="${TRIVY_VERSION:-0.20.0}"
curl "https://raw.githubusercontent.com/aquasecurity/trivy/v${TRIVY_VERSION}/contrib/sarif.tpl" >"${SARIF_TEMPLATE_FILE}"

# generate security report
trivy image --format template --template "@${SARIF_TEMPLATE_FILE}" -o "${SARIF_REPORT_FILE}" "${IMAGE}"

# compress and encode security report
ZIP_FILE="report.gz"
gzip -c "${SARIF_REPORT_FILE}" >"${ZIP_FILE}"

GITHUB_SARIF_FILE="github.sarif"

## there are different flags available based on bsd or gnu base64
if ! base64 -w0 <"${ZIP_FILE}" >"${GITHUB_SARIF_FILE}"; then
  base64 <"${ZIP_FILE}" >"${GITHUB_SARIF_FILE}"
fi

# upload to github
# see https://docs.github.com/en/rest/reference/code-scanning#upload-an-analysis-as-sarif-data

DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

ARGS=(
  "--header"
  "accept:application/vnd.github.v3+json"

  "--raw-field"
  "tool_name=trivy@${APP}"

  "--field"
  "commit_sha=${COMMIT:-${BUILDKITE_COMMIT}}"

  # we can also upload to a pull request to add checks to it directly
  # but I'm doing the simple case for now
  "--field"
  "ref=refs/heads/${BUILDKITE_BRANCH}"

  "--field"
  "checkout_uri=${APP}"

  "--field"
  "started_at=${DATE}"

  "--field"
  "sarif=@${GITHUB_SARIF_FILE}"
)

API_SLUG="/repos/sourcegraph/sourcegraph/code-scanning/sarifs"
gh api "${API_SLUG}" "${ARGS[@]}"

popd
