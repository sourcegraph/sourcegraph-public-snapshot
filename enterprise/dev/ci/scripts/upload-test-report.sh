#!/usr/bin/env bash

xml_file=$1
xml=$(cat "$xml_file")

test_key_variable_name=$2

# escape xml output properly for JSON
quoted_xml="$(echo "$xml" | jq -R -s '.')"
data=$(
  cat <<EOF
{
  "format": "junit",
  "run_env": {
    "CI": "buildkite",
    "key": "$BUILDKITE_BUILD_ID",
    "number": "$BUILDKITE_BUILD_NUMBER",
    "job_id": "$BUILDKITE_JOB_ID",
    "branch": "$BUILDKITE_BRANCH",
    "commit_sha": "$BUILDKITE_COMMIT",
    "message": "$BUILDKITE_MESSAGE",
    "url": "$BUILDKITE_BUILD_URL"
  },
  "data": $quoted_xml
}
EOF
)

TOKEN=$(gcloud secrets versions access latest --secret="$test_key_variable_name" --project="sourcegraph-ci" --quiet)

set +e
echo "$data" | curl \
  --fail \
  --request POST \
  --url https://analytics-api.buildkite.com/v1/uploads \
  --header "Authorization: Token token=\"$TOKEN\";" \
  --header 'Content-Type: application/json' \
  --data-binary @-
curl_exit="$?"
if [ "$curl_exit" -eq 0 ]; then
  echo -e "\n:information_source: Succesfully uploaded test results to Buildkite analytics"
else
  echo -e "\n^^^ +++ :warning: Failed to upload test results to Buildkite analytics"
fi
set -e
