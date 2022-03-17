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

# branchName: BUILDKITE_BRANCH or current git branch
branchName="${BUILDKITE_BRANCH}"

if [ -z "${branchName}" ]; then
  branchName=$(git rev-parse --abbrev-ref HEAD)
fi

# repoUrl: BUILDKITE_PULL_REQUEST_REPO or current git remote `origin` url
# Should be in formats:
# - https://github.com/sourcegraph/sourcegraph
# - git://github.com/sourcegraph/sourcegraph.git
repoUrl="${BUILDKITE_PULL_REQUEST_REPO}"

if [ -z "${repoUrl}" ]; then
  repoUrl=$(git config --get remote.origin.url)
fi

while getopts 'b:r:d' flag; do
  case "${flag}" in
    d) isDeleting="true" ;;
    b) branchName="${OPTARG}" ;;
    r) repoUrl="${OPTARG}" ;;
    *)
      print_usage
      exit 1
      ;;
  esac
done

if [[ "$repoUrl" =~ ^(https|git):\/\/github\.com\/(.*)$ ]]; then
  ownerAndRepo=$(echo "${BASH_REMATCH[2]}" | sed 's/\.git//')
else
  echo "Couldn't find ownerAndRepo"
  exit 1
fi

renderApiKey="${RENDER_COM_API_KEY}"
renderOwnerId="${RENDER_COM_OWNER_ID}"

if [[ -z "${renderApiKey}" || -z "${renderOwnerId}" ]]; then
  echo "RENDER_COM_API_KEY or RENDER_COM_OWNER_ID is not set"
  exit 1
fi

echo "repoUrl: ${repoUrl}"
echo "branchName: ${branchName}"
echo "ownerAndRepo: ${ownerAndRepo}"

prPreviewAppName="sg-web-${branchName}"
pullRequestNumber="${BUILDKITE_PULL_REQUEST}"
githubToken="${GITHUB_TOKEN}"

renderServiceId=$(curl -sS --request GET \
  --url "https://api.render.com/v1/services?limit=1&type=web_service&name=$(urlencode "$prPreviewAppName")" \
  --header 'Accept: application/json' \
  --header "Authorization: Bearer ${renderApiKey}" | jq -r '.[].service.id')

if [ "${isDeleting}" = "true" ]; then
  if [ -z "${renderServiceId}" ]; then
    echo "Render app not found"

    exit 0
  fi

  curl -sSf -o /dev/null --request DELETE \
    --url "https://api.render.com/v1/services/${renderServiceId}" \
    --header 'Accept: application/json' \
    --header "Authorization: Bearer ${renderApiKey}"

  echo "Render app is deleted!"

  exit 0
fi

if [ -z "${renderServiceId}" ]; then
  prPreviewUrl=$(curl -sSf --request POST \
    --url https://api.render.com/v1/services \
    --header 'Accept: application/json' \
    --header 'Content-Type: application/json' \
    --header "Authorization: Bearer ${renderApiKey}" \
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
        \"name\": \"${prPreviewAppName}\",
        \"ownerId\": \"${renderOwnerId}\",
        \"repo\": \"${repoUrl}\",
        \"branch\": \"${branchName}\"
    }
    " | jq -r '.service.serviceDetails.url')
else
  prPreviewUrl=$(curl -sSf --request GET \
    --url "https://api.render.com/v1/services/${renderServiceId}" \
    --header 'Accept: application/json' \
    --header 'Content-Type: application/json' \
    --header "Authorization: Bearer ${renderApiKey}" | jq -r '.serviceDetails.url')
fi

echo "prPreviewUrl: ${prPreviewUrl}"

if [[ -z "${pullRequestNumber}" || -z "${githubToken}" ]]; then
  echo "Pull request number (BUILDKITE_PULL_REQUEST) or github token (GITHUB_TOKEN) is not set, abort updating PR description step"
else
  echo "Update PR description with PR preview app url"

  githubAPIUrl="https://api.github.com/repos/${ownerAndRepo}/pulls/${pullRequestNumber}"

  prBody=$(curl -sSf --request GET \
    --url "${githubAPIUrl}" \
    --user "apikey:${githubToken}" \
    --header 'Accept: application/vnd.github.v3+json' \
    --header 'Content-Type: application/json' | jq -r '.body')

  if [[ "${prBody}" != *"## App preview"* ]]; then
    prBody=$(echo -e "${prBody}\n## App preview:\n- [Link](${prPreviewUrl})\n" | jq -Rs .)

    curl -sSf -o /dev/null --request PATCH \
      --url "${githubAPIUrl}" \
      --user "apikey:${githubToken}" \
      --header 'Accept: application/vnd.github.v3+json' \
      --header 'Content-Type: application/json' \
      --data "{ \"body\": ${prBody} }"
  fi
fi
