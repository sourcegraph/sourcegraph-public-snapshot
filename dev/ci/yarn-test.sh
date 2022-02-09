#!/usr/bin/env bash

set -e

echo "--- yarn in root"
# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --network-timeout 60000

echo "--- generate"
yarn gulp generate

cd "$1"
echo "--- test"

function yarn_test() {
  JEST_JUNIT_OUTPUT_NAME="jest-junit.xml"
  export JEST_JUNIT_OUTPUT_NAME
  JEST_JUNIT_OUTPUT_DIR=$(mktemp -d)
  export JEST_JUNIT_OUTPUT_DIR
  trap 'rm -Rf "$JEST_JUNIT_OUTPUT_DIR"' EXIT

  set +eo pipefail # so we still get the result if the test failed
  local test_exit_code

  # Limit the number of workers to prevent the default of 1 worker per core from
  # causing OOM on the buildkite nodes that have 96 CPUs. 4 matches the CPU limits
  # in infrastructure/kubernetes/ci/buildkite/buildkite-agent/buildkite-agent.Deployment.yaml
  yarn -s run test --maxWorkers 4 --verbose --testResultsProcessor jest-junit

  # Save the test exit code so we can return it after submitting the test run to the analytics.
  test_exit_code="$?"

  set -eo pipefail # resume being strict about errors

  # escape xml output properly for JSON
  local quoted_xml
  quoted_xml="$(jq -R -s '.' "$JEST_JUNIT_OUTPUT_DIR/$JEST_JUNIT_OUTPUT_NAME")"

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

  echo "$data" | curl \
    --request POST \
    --url https://analytics-api.buildkite.com/v1/uploads \
    --header "Authorization: Token token=\"J6giCC8KZFYoVnvgojYu1ESG\";" \
    --header 'Content-Type: application/json' \
    --data-binary @-

  echo -e "\n--- :information_source: Succesfully uploaded test results to Buildkite analytics"

  unest JEST_JUNIT_OUTPUT_DIR
  unest JEST_JUNIT_OUTPUT_NAME
  return "$test_exit_code"
}

yarn_test
