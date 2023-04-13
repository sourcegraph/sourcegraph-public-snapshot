#!/usr/bin/env bash

set -e

# Env variables:
# - RENDER_COM_API_KEY (required)
# - RENDER_COM_OWNER_ID (required)
# - BUILDKITE_BRANCH (optional)
# - BUILDKITE_PULL_REQUEST_REPO (optional)
# - BUILDKITE_PULL_REQUEST (optional)
# - RENDER_PREVIEW_GITHUB_TOKEN (optional)

print_usage() {
  echo "Usage: [ -b BRANCH_NAME ] [ -r REPO_URL ] [ -d ]" 1>&2
  echo "-b: GitHub branch name" 1>&2
  echo "-r: GitHub repository url" 1>&2
  echo "-d: Use this flag to delete preview apps" 1>&2
}

urlencode() {
  echo "$1" | curl -Gso /dev/null -w "%{url_effective}" --data-urlencode @- "" | cut -c 3- | sed -e 's/%0A//'
}

# branch_name: BUILDKITE_BRANCH or current git branch
branch_name="${BUILDKITE_BRANCH}"

if [ -z "${branch_name}" ]; then
  branch_name=$(git rev-parse --abbrev-ref HEAD)
fi

# repo_url: BUILDKITE_PULL_REQUEST_REPO or current git remote `origin` url
# Should be in formats:
# - https://github.com/sourcegraph/sourcegraph
# - git://github.com/sourcegraph/sourcegraph.git
repo_url="${BUILDKITE_PULL_REQUEST_REPO}"

if [ -z "${repo_url}" ]; then
  repo_url=$(git config --get remote.origin.url)
fi

while getopts 'b:r:d' flag; do
  case "${flag}" in
    d) is_deleting="true" ;;
    b) branch_name="${OPTARG}" ;;
    r) repo_url="${OPTARG}" ;;
    *)
      print_usage
      exit 1
      ;;
  esac
done

# Get `{owner}/{repo}` part from GitHub repository url
if [[ "$repo_url" =~ ^(https://|git://|git@)github\.com(:|\/)(.*)$ ]]; then
  owner_and_repo="${BASH_REMATCH[3]//\.git/}"
else
  echo "Couldn't find owner_and_repo"
  exit 1
fi

render_api_key="${RENDER_COM_API_KEY}"
render_owner_id="${RENDER_COM_OWNER_ID}"
pr_number="${BUILDKITE_PULL_REQUEST}"
github_api_key="${RENDER_PREVIEW_GITHUB_TOKEN}"

if [[ -z "${render_api_key}" || -z "${render_owner_id}" ]]; then
  echo "RENDER_COM_API_KEY or RENDER_COM_OWNER_ID is not set"
  exit 1
fi

# App name to show on render.com dashboard and use to create
# default url: https://<app_name_slug>.onrender.com
pr_preview_app_name="sg-web-${branch_name}"

# Get service id of preview app on render.com with app name (if exists)
renderServiceId=$(curl -sSf --request GET \
  --url "https://api.render.com/v1/services?limit=1&type=web_service&name=$(urlencode "$pr_preview_app_name")" \
  --header 'Accept: application/json' \
  --header "Authorization: Bearer ${render_api_key}" | jq -r '.[].service.id')

# Delete preview app with `-d` flag set
if [ "${is_deleting}" = "true" ]; then
  if [ -z "${renderServiceId}" ]; then
    echo "Render app not found"

    exit 0
  fi

  curl -sSf -o /dev/null --request DELETE \
    --url "https://api.render.com/v1/services/${renderServiceId}" \
    --header 'Accept: application/json' \
    --header "Authorization: Bearer ${render_api_key}"

  echo "Render app is deleted!"

  exit 0
fi

# Create PR app if it hasn't existed yet and get the app url
if [ -z "${renderServiceId}" ]; then
  echo "Creating new pr preview app..."

  # New app is created with following envs
  # - ENTERPRISE=1
  # - NODE_ENV=production
  # - PORT=3080 // render.com uses this env for mapping default https port
  # - ENTERPRISE=1
  # - SOURCEGRAPH_API_URL=https://k8s.sgdev.org
  # - WEBPACK_SERVE_INDEX=true

  pr_preview_url=$(curl -sSf --request POST \
    --url https://api.render.com/v1/services \
    --header 'Accept: application/json' \
    --header 'Content-Type: application/json' \
    --header "Authorization: Bearer ${render_api_key}" \
    --data "
    {
        \"autoDeploy\": \"yes\",
        \"envVars\": [
            {
                \"key\": \"ENTERPRISE\",
                \"value\": \"1\"
            },
            {
                \"key\": \"NODE_ENV\",
                \"value\": \"production\"
            },
            {
                \"key\": \"PORT\",
                \"value\": \"3080\"
            },
            {
                \"key\": \"SOURCEGRAPH_API_URL\",
                \"value\": \"https://k8s.sgdev.org\"
            },
            {
                \"key\": \"WEBPACK_SERVE_INDEX\",
                \"value\": \"true\"
            }
        ],
        \"serviceDetails\": {
            \"pullRequestPreviewsEnabled\": \"no\",
            \"envSpecificDetails\": {
                \"buildCommand\": \"source dev/ci/render-preview-install.sh && dev/ci/pnpm-build.sh client/web\",
                \"startCommand\": \"source dev/ci/render-preview-install.sh && pnpm --filter @sourcegraph/web serve:prod\"
            },
            \"numInstances\": 1,
            \"plan\": \"starter\",
            \"region\": \"oregon\",
            \"env\": \"node\"
        },
        \"type\": \"web_service\",
        \"name\": \"${pr_preview_app_name}\",
        \"ownerId\": \"${render_owner_id}\",
        \"repo\": \"${repo_url}\",
        \"branch\": \"${branch_name}\"
    }
    " | jq -r '.service.serviceDetails.url')
else
  echo "Found preview id: ${renderServiceId}, getting preview url..."

  pr_preview_url=$(curl -sSf --request GET \
    --url "https://api.render.com/v1/services/${renderServiceId}" \
    --header 'Accept: application/json' \
    --header 'Content-Type: application/json' \
    --header "Authorization: Bearer ${render_api_key}" | jq -r '.serviceDetails.url')
fi

echo "Preview url: ${pr_preview_url}"

if [[ -n "${github_api_key}" && -n "${pr_number}" && "${pr_number}" != "false" ]]; then

  # GitHub pull request number and GitHub api token are set
  # Appending `App Preview` section into PR description if it hasn't existed yet
  github_pr_api_url="https://api.github.com/repos/${owner_and_repo}/pulls/${pr_number}"

  pr_description=$(curl -sSf --request GET \
    --url "${github_pr_api_url}" \
    --user "apikey:${github_api_key}" \
    --header 'Accept: application/vnd.github.v3+json' \
    --header 'Content-Type: application/json' | jq -r '.body')

  if [[ "${pr_description}" != *"## App preview"* ]]; then
    echo "Updating PR #${pr_number} in ${owner_and_repo} description"

    pr_description=$(printf '%s\n\n' "${pr_description}" \
      "## App preview:" \
      "- [Web](${pr_preview_url}/search)" \
      "Check out the [client app preview documentation](https://docs.sourcegraph.com/dev/how-to/client_pr_previews) to learn more." |
      jq -Rs .)

    curl -sSf -o /dev/null --request PATCH \
      --url "${github_pr_api_url}" \
      --user "apikey:${github_api_key}" \
      --header 'Accept: application/vnd.github.v3+json' \
      --header 'Content-Type: application/json' \
      --data "{ \"body\": ${pr_description} }"
  else
    echo "PR #${pr_number} in ${owner_and_repo} description already has \"App preview\" section"
  fi
fi
