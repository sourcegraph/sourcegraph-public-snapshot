#!/usr/bin/env bash

set -eu -o pipefail

# This script gets Chromatic storybook preview link from `pnpm chromatic [...]` output
# to add Storybook link into `App preview` section of PR description.
# Ex: pnpm chromatic --exit-zero-on-changes --exit-once-uploaded | ./dev/ci/post-chromatic.sh
# Env variables:
# - BUILDKITE_PULL_REQUEST_REPO
# - BUILDKITE_PULL_REQUEST
# - RENDER_PREVIEW_GITHUB_TOKEN

exit_status=$?
chromatic_publish_output=$(</dev/stdin)

if [ $exit_status -eq 0 ]; then
  echo "$chromatic_publish_output"
else
  echo "🧨 EXIT CODE ${exit_status}"
  echo "$chromatic_publish_output"
  exit 1
fi

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

if [[ -n "${github_api_key}" && -n "${pr_number}" && "${pr_number}" != "false" ]]; then
  github_pr_comments_api_url="https://api.github.com/repos/${owner_and_repo}/issues/${pr_number}/comments"

  app_preview_comment_id=$(curl -sSf --request GET \
    --url "${github_pr_comments_api_url}" \
    --user "apikey:${github_api_key}" \
    --header 'Accept: application/vnd.github.v3+json' \
    --header 'Content-Type: application/json' | jq '.[] | select(.body | contains("📖 [Storybook live preview]")) | .id')

  app_preview_comment_body=$(printf '%s\n' \
    "📖 [Storybook live preview](${chromatic_storybook_url})" | jq -Rs .)

  if [[ -z "${app_preview_comment_id}" ]]; then
    echo "Adding new App preview comment to PR #${pr_number} in ${owner_and_repo}"

    curl -sSf -o /dev/null --request POST \
      --url "${github_pr_comments_api_url}" \
      --user "apikey:${github_api_key}" \
      --header 'Accept: application/vnd.github.v3+json' \
      --header 'Content-Type: application/json' \
      --data "{ \"body\": ${app_preview_comment_body} }"
  else
    echo "Updating App preview comment (id: ${app_preview_comment_id}) in PR #${pr_number} in ${owner_and_repo}"

    curl -sSf -o /dev/null --request PATCH \
      --url "https://api.github.com/repos/${owner_and_repo}/issues/comments/${app_preview_comment_id}" \
      --user "apikey:${github_api_key}" \
      --header 'Accept: application/vnd.github.v3+json' \
      --header 'Content-Type: application/json' \
      --data "{ \"body\": ${app_preview_comment_body} }"
  fi
fi
