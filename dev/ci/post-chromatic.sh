#!/usr/bin/env bash

set -eu -o pipefail

# Env variables:
# - BUILDKITE_PULL_REQUEST_REPO
# - BUILDKITE_PULL_REQUEST
# - RENDER_PREVIEW_GITHUB_TOKEN

chromatic_publish_output=$(</dev/stdin)

# Chromatic preview url from Chromatic publish output (`-m 1` for getting first match)
chromatic_storybok_url=$(echo "$chromatic_publish_output" | grep -oh -m 1 "https:\/\/[[:alnum:]]*\-[[:alnum:]]*\.chromatic\.com")

if [ -z "${chromatic_storybok_url}" ]; then
  echo "Couldn't find Chromatic preview url"
  exit 1
fi

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
  # Appending `App Preview` section into PR description if it hasn't existed yet
  github_pr_api_url="https://api.github.com/repos/${owner_and_repo}/pulls/${pr_number}"

  pr_description=$(curl -sSf --request GET \
    --url "${github_pr_api_url}" \
    --user "apikey:${github_api_key}" \
    --header 'Accept: application/vnd.github.v3+json' \
    --header 'Content-Type: application/json' | jq -r '.body')

  # Assume Chromatic publish finishes after Render pr preview job
  if [[ "${pr_description}" == *"## App preview"* ]]; then
    echo "Updating PR #${pr_number} in ${owner_and_repo} description"

    pr_description=$(echo "$pr_description" | sed -e '/\[Link\](https:\/\/.*.onrender.com).*/a\
- [Storybook](https://5f0f381c0e50750022dc6bf7-fayzipwrry.chromatic.com)')

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
