#!/usr/bin/env bash

set -e

echo -e "--- :construction: cloning Semgrep rules\n"

# clone the semgrep repo rules
gh repo clone sourcegraph/security-semgrep-rules

echo -e "--- :lock: running Semgrep scan\n"

set -x

# verify security-semgrep-rules/semgrep-rules/ directory is present or print error
if [ ! -d "security-semgrep-rules/semgrep-rules/" ]; then
  echo ":red_circle: Semgrep rules directory not found. Reachout to security team at #discuss-security for support :red_circle:"
fi

# run semgrep
semgrep ci -f security-semgrep-rules/semgrep-rules/ --metrics=off --oss-only --suppress-errors --sarif -o results.sarif --exclude='semgrep-rules' --baseline-commit main

echo -e "--- :rocket: reporting Scan Results to GitHub\n"

# upload SARIF results to code scanning API
encoded_sarif=$(gzip -c results.sarif | base64 -w0)

# upload SARIF results to code scanning API
gh api \
  --method POST \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  /repos/sourcegraph/sourcegraph/code-scanning/sarifs \
  -f commit_sha="$BUILDKITE_COMMIT" \
  -f ref="refs/pull/$BUILDKITE_PULL_REQUEST/head" \
  -f sarif="$encoded_sarif"

echo -e "--- :white_check_mark: Semgrep Scan job is complete\n"
