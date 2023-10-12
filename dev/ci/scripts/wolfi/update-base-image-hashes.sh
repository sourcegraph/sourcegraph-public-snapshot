#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# Update hashes for all base images
go run ./dev/sg wolfi update-hashes

# DEBUG: Print oci_deps
cat dev/oci_deps.bzl

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

# Run gh
gh --version

BRANCH_NAME="wolfi-autoupdate/main"
PR_TITLE="Update Wolfi base images to latest"
TIMESTAMP=$(TZ=UTC date "+%Y-%m-%d %H:%M:%S %z")

# Commit changes to dev/oci-deps.bzl
# Delete branch if it exists; catch status code if not
git branch -D "${BRANCH_NAME}" || :
git checkout -b "${BRANCH_NAME}"
git add dev/oci_deps.bzl
git commit -m "Automatically update Wolfi base image hashes at ${TIMESTAMP}"
# git remote set-url token-origin https://sg-test:${GH_TOKEN}@github.com/sourcegraph/sourcegraph.git
# git push --force -u token-origin "${BRANCH_NAME}"
echo "Successfully commited changes and pushed to branch ${BRANCH_NAME}"

# Check if an update PR already exists

if gh pr list --head "${BRANCH_NAME}" --state open | grep "${PR_TITLE}"; then
  echo "A pull request already exists - no action required"
else
  # If not, create a new PR from the branch foobar-day
  # gh pr create --title "foobar" --head "${BRANCH_NAME}" --base main --body "Automatically generated PR to update Wolfi base images to the latest hashes"
  echo "Would have created a new pull request from branch ${BRANCH_NAME} with title ${PR_TITLE}"
fi
