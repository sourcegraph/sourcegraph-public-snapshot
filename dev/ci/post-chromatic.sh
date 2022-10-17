#!/usr/bin/env bash

set -eu -o pipefail

# This script gets Chromatic storybook preview link from `yarn chromatic [...]` output
# to add Storybook link into `App preview` section of PR description.
# Ex: yarn chromatic --exit-zero-on-changes --exit-once-uploaded | ./dev/ci/post-chromatic.sh
# Env variables:
# - BUILDKITE_PULL_REQUEST_REPO
# - BUILDKITE_PULL_REQUEST
# - RENDER_PREVIEW_GITHUB_TOKEN

chromatic_publish_output=$(</dev/stdin)

echo "$chromatic_publish_output"

# Chromatic preview url from Chromatic publish output (`-m 1` for getting first match)
chromatic_storybook_url=$(echo "$chromatic_publish_output" | grep -oh -m 1 "https:\/\/[[:alnum:]]*\-[[:alnum:]]*\.chromatic\.com")

if [ -z "${chromatic_storybook_url}" ]; then
  echo "Couldn't find Chromatic preview url"
  exit 1
fi

echo "Found $chromatic_storybook_url"

# repo_url: BUILDKITE_PULL_REQUEST_REPO or current git remote `origin` url
# Should be in formats:
# - https://github.com/sourcegraph/sourcegraph
# - git://github.com/sourcegraph/sourcegraph.git
repo_url="${BUILDKITE_PULL_REQUEST_REPO}"
pr_number="${BUILDKITE_PULL_REQUEST}"
github_api_key="${RENDER_PREVIEW_GITHUB_TOKEN}"

# Get `{owner}/{repo}` part from GitHub repository url
if [[ "$repo_url" =~ ^(https|git):\/\/github\.com\/(.*)$ ]]; then
  owner_and_repo="${BASH_REMATCH[2]//\.git/}"
else
  echo "Couldn't find owner_and_repo"
  exit 1
fi

if [[ -n "${github_api_key}" && -n "${pr_number}" ]]; then
  # GitHub pull request number and GitHub api token are set
  # Appending Storybook link into App preview section
  github_pr_api_url="https://api.github.com/repos/${owner_and_repo}/pulls/${pr_number}"

  pr_description=$(curl -sSf --request GET \
    --url "${github_pr_api_url}" \
    --user "apikey:${github_api_key}" \
    --header 'Accept: application/vnd.github.v3+json' \
    --header 'Content-Type: application/json' | jq -r '.body')

  # Assume Chromatic publish finishes after Render PR preview job
  if [[ "${pr_description}" == *"## App preview"* ]]; then
    echo "Updating PR #${pr_number} in ${owner_and_repo} description"

    # Check if Storybook link exists for adding new link or replacing existing one
    if [[ "$pr_description" =~ \[Storybook\]\(https:\/\/[[:alnum:]]*\-[[:alnum:]]*\.chromatic\.com\) ]]; then
      pr_description=$(echo "$pr_description" | sed -e "s|\[Storybook\](https:\/\/[[:alnum:]]*\-[[:alnum:]]*\.chromatic\.com)|[Storybook](${chromatic_storybook_url})|" | jq -Rs .)
    else
      pr_description=$(echo "$pr_description" | sed -e "s|\[[[:alpha:]]*\](https:\/\/.*.onrender.com)|&\n- [Storybook](${chromatic_storybook_url})|" | jq -Rs .)
    fi

    curl -sSf -o /dev/null --request PATCH \
      --url "${github_pr_api_url}" \
      --user "apikey:${github_api_key}" \
      --header 'Accept: application/vnd.github.v3+json' \
      --header 'Content-Type: application/json' \
      --data "{ \"body\": ${pr_description} }"
  else
    echo "Couldn't find \"App preview\" section in description of PR #${pr_number} in ${owner_and_repo}"
  fi
fi
