#!/usr/bin/env bash

# Env variables:
# - RENDER_COM_API_KEY (required)
# - RENDER_COM_OWNER_ID (required)
# - BUILDKITE_BRANCH (optional)
# - BUILDKITE_PULL_REQUEST_REPO (optional)
# - BUILDKITE_PULL_REQUEST (optional)
# - GITHUB_TOKEN (optional)

echo "--- Render.com PR Preview"

print_usage() {
  echo "Usage: $0 [ -b BRANCH_NAME ] [ -r REPO_URL ] [ -d ]" 1>&2
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

if [[ "$repo_url" =~ ^(https|git):\/\/github\.com\/(.*)$ ]]; then
  owner_and_repo="${BASH_REMATCH[2]//\.git/}"
else
  echo "Couldn't find owner_and_repo"
  exit 1
fi

render_api_key="${RENDER_COM_API_KEY}"
render_owner_id="${RENDER_COM_OWNER_ID}"
pr_number="${BUILDKITE_PULL_REQUEST}"
github_api_key="${GITHUB_TOKEN}"

if [[ -z "${render_api_key}" || -z "${render_owner_id}" ]]; then
  echo "RENDER_COM_API_KEY or RENDER_COM_OWNER_ID is not set"
  exit 1
fi

echo "repo_url: ${repo_url}"
echo "branch_name: ${branch_name}"
echo "owner_and_repo: ${owner_and_repo}"

pr_preview_app_name="sg-web-${branch_name}"

renderServiceId=$(curl -sS --request GET \
  --url "https://api.render.com/v1/services?limit=1&type=web_service&name=$(urlencode "$pr_preview_app_name")" \
  --header 'Accept: application/json' \
  --header "Authorization: Bearer ${render_api_key}" | jq -r '.[].service.id')

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

if [ -z "${renderServiceId}" ]; then
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
                \"key\": \"SOURCEGRAPHDOTCOM_MODE\",
                \"value\": \"false\"
            },
            {
                \"key\": \"WEBPACK_SERVE_INDEX\",
                \"value\": \"true\"
            }
        ],
        \"serviceDetails\": {
            \"pullRequestPreviewsEnabled\": \"no\",
            \"envSpecificDetails\": {
                \"buildCommand\": \"yarn install && dev/ci/yarn-build.sh client/web\",
                \"startCommand\": \"yarn workspace @sourcegraph/web serve:prod\"
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
  pr_preview_url=$(curl -sSf --request GET \
    --url "https://api.render.com/v1/services/${renderServiceId}" \
    --header 'Accept: application/json' \
    --header 'Content-Type: application/json' \
    --header "Authorization: Bearer ${render_api_key}" | jq -r '.serviceDetails.url')
fi

echo "pr_preview_url: ${pr_preview_url}"

if [[ -z "${pr_number}" || -z "${github_api_key}" ]]; then
  echo "Pull request number (BUILDKITE_PULL_REQUEST) or github token (GITHUB_TOKEN) is not set, abort updating PR description step"
else
  echo "Update PR description with PR preview app url"

  github_api_url="https://api.github.com/repos/${owner_and_repo}/pulls/${pr_number}"

  pr_description=$(curl -sSf --request GET \
    --url "${github_api_url}" \
    --user "apikey:${github_api_key}" \
    --header 'Accept: application/vnd.github.v3+json' \
    --header 'Content-Type: application/json' | jq -r '.body')

  if [[ "${pr_description}" != *"## App preview"* ]]; then
    pr_description=$(echo -e "${pr_description}\n## App preview:\n- [Link](${pr_preview_url})\n" | jq -Rs .)

    curl -sSf -o /dev/null --request PATCH \
      --url "${github_api_url}" \
      --user "apikey:${github_api_key}" \
      --header 'Accept: application/vnd.github.v3+json' \
      --header 'Content-Type: application/json' \
      --data "{ \"body\": ${pr_description} }"
  fi
fi
