#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
set -ex

URL="${1:-"http://localhost:7080"}"

function integration_test() {
  MOCHA_JUNIT_OUTPUT_DIR=$(mktemp -d)
  export MOCHA_JUNIT_OUTPUT_DIR
  MOCHA_FILE="$MOCHA_JUNIT_OUTPUT_DIR/mocha-junit.xml"
  export MOCHA_FILE
  trap 'rm -Rf "$MOCHA_JUNIT_OUTPUT_DIR"' EXIT

  set +eo pipefail # so we still get the result if the test failed
  local test_exit_code

  echo "--- yarn run test-e2e"
  env SOURCEGRAPH_BASE_URL="$URL" PERCY_ON=true ./node_modules/.bin/percy exec -- yarn run cover-e2e --reporter mocha-junit-reporter

  # Save the test exit code so we can return it after submitting the test run to the analytics.
  test_exit_code="$?"

  set -eo pipefail # resume being strict about errors

  # escape xml output properly for JSON
  set +x
  local quoted_xml
  quoted_xml="$(jq -R -s '.' "$MOCHA_FILE")"

  local data
  data=$(
    cat <<EOF
{
  "format": "junit",
  "run_env": {
    "CI": "buildkite",
    "key": "$BUILDKITE_BUILD_ID",
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

  echo -e "\n--- :information_source: Uploading test results to Buildkite analytics"
  set +e
  echo "$data" | curl \
    --fail \
    --request POST \
    --url https://analytics-api.buildkite.com/v1/uploads \
    --header "Authorization: Token token=\"$BUILDKITE_ANALYTICS_FRONTEND_E2E_TEST_SUITE_API_KEY\";" \
    --header 'Content-Type: application/json' \
    --data-binary @-
  local curl_exit="$?"
  if [ "$curl_exit" -eq 0 ]; then
    echo -e "\n--- :information_source: Succesfully uploaded test results to Buildkite analytics"
  else
    echo -e "\n^^^ +++ :warning: Failed to upload test results to Buildkite analytics"
  fi
  set -e

  unset MOCHA_JUNIT_OUTPUT_DIR
  unset MOCHA_FILE
  set -x

  return "$test_exit_code"
}

BUILDKITE_ANALYTICS_FRONTEND_E2E_TEST_SUITE_API_KEY=$(gcloud secrets versions access latest --secret="BUILDKITE_ANALYTICS_FRONTEND_E2E_TEST_SUITE_API_KEY" --project="sourcegraph-ci" --quiet)

integration_test

echo "--- coverage"
yarn nyc report -r json
# Upload the coverage under the "e2e" flag (toggleable in the CodeCov UI)
./dev/ci/codecov.sh -F e2e
