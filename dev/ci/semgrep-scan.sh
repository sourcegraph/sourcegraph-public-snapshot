#!/usr/bin/env bash

set -e

echo -e "--- :construction: cloning Semgrep rules\n"

# clone the semgrep repo rules
gh repo clone sourcegraph/security-semgrep-rules

echo -e "--- :lock::semgrep: running Semgrep scan\n"

set -x

# verify security-semgrep-rules/semgrep-rules/ directory is present or print error
if [ ! -d "security-semgrep-rules/semgrep-rules/" ]; then
  echo ":red_circle: Semgrep rules directory not found. Reachout to security team at #discuss-security for support :red_circle:"
fi

CODE_SCANNING_ENABLED="false"

is_code_scanning_enabled() {
  local repo="$1"

  error=$(gh api "repos/$repo/code-scanning/alerts" || true)
  if echo "$error" | grep -q "Advanced Security must be enabled"; then
    CODE_SCANNING_ENABLED="false"
  elif echo "$error" | grep -q "Not Found"; then
    CODE_SCANNING_ENABLED="false"
  else
    CODE_SCANNING_ENABLED="true"
  fi
}

FULL_REPO_NAME="${CI_REPO_OWNER}/${CI_REPO_NAME}"
is_code_scanning_enabled "$FULL_REPO_NAME"

echo $CODE_SCANNING_ENABLED
# run semgrep scan on changeset using CI subcommand
# || true is used to prevent build from failing if semgrep scan reports on blocking findings
# reference: https://semgrep.dev/docs/semgrep-ci/configuring-blocking-and-errors-in-ci/#configuration-options-for-blocking-findings-and-errors
if [ "$CODE_SCANNING_ENABLED" = "true" ]; then
  semgrep ci -f 'security-semgrep-rules/semgrep-rules/' --metrics=off --oss-only --sarif -o results.sarif --exclude='semgrep-rules' --baseline-commit "$(git merge-base main HEAD)" || true
else
  semgrep ci -f 'security-semgrep-rules/semgrep-rules/' --metrics=off --oss-only --json -o result.json --exclude='semgrep-rules' --baseline-commit "$(git merge-base main HEAD)" || true
fi

if [ "$CODE_SCANNING_ENABLED" = "true" ]; then
  echo -e "--- :rocket: reporting scan results to GitHub\n"

  # encode SARIF results to code scanning API
  encoded_sarif=$(gzip -c results.sarif | base64 -w0)

  # upload SARIF results to code scanning API
  if [ "$BUILDKITE_PULL_REQUEST" = "false" ]; then
    ref="refs/heads/${BUILDKITE_BRANCH}"
    if [[ -n "${BUILDKITE_TAG}" ]]; then
      ref="refs/tags/${BUILDKITE_TAG}"
    fi

    gh api \
      --method POST \
      -H "Accept: application/vnd.github+json" \
      -H "X-GitHub-Api-Version: 2022-11-28" \
      /repos/sourcegraph/sourcegraph/code-scanning/sarifs \
      -f commit_sha="$BUILDKITE_COMMIT" \
      -f ref="${ref}" \
      -f sarif="$encoded_sarif" \
      -f tool_name="ci semgrep"
  else
    gh api \
      --method POST \
      -H "Accept: application/vnd.github+json" \
      -H "X-GitHub-Api-Version: 2022-11-28" \
      /repos/sourcegraph/sourcegraph/code-scanning/sarifs \
      -f commit_sha="$BUILDKITE_COMMIT" \
      -f ref="refs/pull/$BUILDKITE_PULL_REQUEST/head" \
      -f sarif="$encoded_sarif" \
      -f tool_name="ci semgrep"
  fi
else
  echo -e "--- :lock::semgrep: processing semgrep results\n"
  # set environment variables to support post-processing script
  export LATEST_COMMIT_SHA=$BUILDKITE_COMMIT
  export GITHUB_PULL_REQUEST_NUMBER=$BUILDKITE_PULL_REQUEST
  export GITHUB_REPOSITORY=$CI_REPO_NAME

  cp result.json security-semgrep-rules/scripts
  cd security-semgrep-rules/scripts
  go mod download
  # this script is located in security-semgrep-rules/scripts repo
  go run main.go
fi

echo -e "--- :white_check_mark::semgrep: Semgrep scan job is complete\n"
