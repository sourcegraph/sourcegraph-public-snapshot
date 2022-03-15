#!/usr/bin/env bash

echo "--- Render.com PR Preview"

curl -sS --request POST \
     --url https://api.render.com/v1/services \
     --header 'Accept: application/json' \
     --header 'Content-Type: application/json' \
     --header "Authorization: Bearer $RENDER_COM_API_KEY" \
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
     \"name\": \"sg-web-$GITHUB_BRANCH\",
     \"ownerId\": \"$RENDER_COM_OWNER_ID\",
     \"repo\": \"$RENDER_COM_REPO\",
     \"branch\": \"$GITHUB_BRANCH\"
}
"

prPreviewUrl=$( curl -sS --request GET \
     --url 'https://api.render.com/v1/services?limit=1' \
     --data-urlencode "name=sg-web-$GITHUB_BRANCH" \
     --header 'Accept: application/json' \
     --header "Authorization: Bearer $RENDER_COM_API_KEY" | jq -r '.[].service.serviceDetails.url' )

prBody=$( curl -sS --request GET \
     --url "https://api.github.com/repos/$GITHUB_REPOSITORY/pulls/$GITHUB_PR" \
     --user "apikey:$GITHUB_TOKEN" \
     --header 'Accept: application/vnd.github.v3+json' \
     --header 'Content-Type: application/json' | jq -r '.body' )

prBody=$(echo "${prBody}
## App preview:
  - [Link](${prPreviewUrl})
" | jq -Rs .)

curl -sS --request PATCH \
     --url "https://api.github.com/repos/$GITHUB_REPOSITORY/pulls/$GITHUB_PR" \
     --user "apikey:$GITHUB_TOKEN" \
     --header 'Accept: application/vnd.github.v3+json' \
     --header 'Content-Type: application/json' \
     --data "
{
     \"body\": ${prBody}
}"
