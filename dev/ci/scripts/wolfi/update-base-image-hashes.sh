#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# Update hashes for all base images
go run ./dev/sg wolfi update-hashes
# Print diff
git diff dev/oci_deps.bzl

# Temporary: Install GitHub CLI
ghtmpdir=$(mktemp -d -t github-cli.XXXXXXXX)
curl -L https://github.com/cli/cli/releases/download/v2.36.0/gh_2.36.0_linux_amd64.tar.gz -o "${ghtmpdir}/gh.tar.gz"
# From https://github.com/cli/cli/releases/download/v2.36.0/gh_2.36.0_checksums.txt
expected_hash="29ed6c04931e6ac8a5f5f383411d7828902fed22f08b0daf9c8ddb97a89d97ce"
actual_hash=$(sha256sum "${ghtmpdir}/gh.tar.gz" | cut -d ' ' -f 1)
if [ "$expected_hash" = "$actual_hash" ]; then
  echo "Hashes match"
else
  echo "Error - hashes do not match!"
  exit 1
fi
tar -xzf "${ghtmpdir}/gh.tar.gz" -C "${ghtmpdir}/"
cp "${ghtmpdir}/gh_2.36.0_linux_amd64/bin/gh" "/usr/local/bin/"
# Test gh
gh --version

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
# git remote set-url token-origin https://sg-test:${GH_TOKEN}@github.com/sourcegraph/sourcegraph.git
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
