#!/usr/bin/env bash

set -e

echo -e "--- :construction: cloning Semgrep rules\n"

# clone the semgrep repo rules
gh repo clone sourcegraph/security-semgrep-rules

echo -e "--- :lock: running Semgrep scan\n"

set -x

# run semgrep
semgrep ci -f semgrep-rules/ --metrics=off --oss-only --suppress-errors --sarif -o results.sarif --exclude='semgrep-rules'

echo -e "--- :rocket: uploading Semgrep scan results\n"

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
