#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -euxo pipefail

# HACK: For whatever reason, we hard-code our
# own, different github token in our pipeline's env settings.
# This causes authentication requests to fail. Setting this
# here is a workaround until we can dedicate time to
# figure out why this is the case.
# TODO: @sourcegraph/distribution
export GITHUB_TOKEN="${GITHUB_TOKEN_COPY:-DEPLOY_SOURCEGRAPH_GITHUB_TOKEN}"

COMMIT="${BUILDKITE_COMMIT}"

API_SLUG="repos/sourcegraph/sourcegraph/commits"
function get_branch_tip() {
  local ref="$1"

  # https://docs.github.com/en/rest/reference/repos#list-commits
  gh api "${API_SLUG}?sha=${ref}&per_page=1" --jq '.[].sha'
}

REF="main"
tip_of_main="$(get_branch_tip ${REF})"

[[ "$tip_of_main" == "$COMMIT" ]]
