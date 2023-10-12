#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# Update hashes for all base images
go run ./dev/sg wolfi update-hashes
# Print diff
git diff dev/oci_deps.bzl

# Git and GitHub config
BRANCH_NAME="wolfi-autoupdate/main"
TIMESTAMP=$(TZ=UTC date "+%Y-%m-%d %H:%M:%S %z")
PR_TITLE="Update Wolfi base images to latest"
# PR_REVIEWER="sourcegraph/security"
PR_LABELS="SSDLC,wolfi-auto-update"
PR_BODY="Automatically generated PR to update Wolfi base images to the latest hashes.
## Test Plan
- CI build verifies image functionality"

# Commit changes to dev/oci-deps.bzl
# Delete branch if it exists; catch status code if not
git branch -D "${BRANCH_NAME}" || :
git checkout -b "${BRANCH_NAME}"
git add dev/oci_deps.bzl
git commit -m "Automatically update Wolfi base image hashes at ${TIMESTAMP}"
git push --force -u origin "${BRANCH_NAME}"
echo ":git: Successfully commited changes and pushed to branch ${BRANCH_NAME}"

# Check if an update PR already exists
if gh pr list --head "${BRANCH_NAME}" --state open | grep -q "${PR_TITLE}"; then
  echo ":github: A pull request already exists - no action required"
else
  # If not, create a new PR from the branch foobar-day
  # TODO: Once validated add '--reviewer "${PR_REVIEWER}"'
  gh pr create --title "${PR_TITLE}" --head "${BRANCH_NAME}" --base main --body "${PR_BODY}" --label "${PR_LABELS}"
  echo ":github: Created a new pull request from branch '${BRANCH_NAME}' with title '${PR_TITLE}'"
fi
