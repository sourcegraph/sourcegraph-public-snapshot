#!/usr/bin/env bash

# Run `sg wolfi lock` to update all package lockfiles for Wolfi base images.
# Push a new branch to GitHub, and open a PR.
# Can be run from any base branch, and will create an appropriate PR.

set -exu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

echo "~~~ :aspect: :stethoscope: Agent Health check"
/etc/aspect/workflows/bin/agent_health_check

aspectRC="/tmp/aspect-generated.bazelrc"
rosetta bazelrc >"$aspectRC"
export BAZELRC="$aspectRC"

echo "~~~ :lock: Run sg wolfi lock"

buildkite-agent artifact download sg . --step bazel-prechecks
chmod +x ./sg

# Update hashes for all base images
./sg wolfi lock

echo "~~~ :github: Commit changes and open PR"

# Print git status
echo "Running git status"
git status

# Git and GitHub config
BRANCH_NAME="wolfi-auto-update/${BUILDKITE_BRANCH}"
TIMESTAMP=$(TZ=UTC date "+%Y-%m-%d %H:%M:%S UTC")
PR_TITLE="security: Auto-update package lockfiles for Sourcegraph base images"
# PR_REVIEWER="sourcegraph/security"
PR_LABELS="SSDLC,security-auto-update,security-auto-update/images"
PR_BODY="Automatically generated PR to update package lockfiles for Sourcegraph base images.

Built from Buildkite run [#${BUILDKITE_BUILD_NUMBER}](https://buildkite.com/sourcegraph/sourcegraph/builds/${BUILDKITE_BUILD_NUMBER}).
## Test Plan
- CI build verifies image functionality"

# Ensure git author details are correct
git config --local user.email \"buildkite@sourcegraph.com\"
git config --local user.name \"Buildkite\"

# Commit changes to dev/oci-deps.bzl
# Delete branch if it exists; catch status code if not
echo "Deleting branch ${BRANCH_NAME} if it exists"
git branch -D "${BRANCH_NAME}" || true
echo "Switching to new branch ${BRANCH_NAME}"
git switch -c "${BRANCH_NAME}"
echo "Git add lockfiles"
git add wolfi-images/*.lock.json
echo "Committing changes"
git commit -m "Auto-update package lockfiles for Wolfi base images at ${TIMESTAMP}"
echo "Git log"
git log -n 1
echo "Pushing changes"
git push --force -u origin "${BRANCH_NAME}"
echo "Successfully commited changes and pushed to branch ${BRANCH_NAME}"

# Generate a wrapper script for GitHub CLI tool
bazel --bazelrc=${aspectRC} run --script_path=gh.sh --noshow_progress --noshow_loading_progress //dev/tools:gh

# Check if an update PR already exists
if ./gh.sh pr list --head "${BRANCH_NAME}" --state open | grep -q "${PR_TITLE}"; then
  echo "A pull request already exists - editing it"
  ./gh.sh pr edit "${BRANCH_NAME}" --body "${PR_BODY}"
else
  # If not, create a new PR from the branch
  ./gh.sh pr create --title "${PR_TITLE}" --head "${BRANCH_NAME}" --base "${BUILDKITE_BRANCH}" --body "${PR_BODY}" --label "${PR_LABELS}"
  echo "Created a new pull request from branch '${BRANCH_NAME}' with title '${PR_TITLE}'"
fi
