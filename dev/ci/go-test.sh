#!/usr/bin/env bash

set -euo pipefail

function usage {
  cat <<EOF
Usage: go-test.sh [only|exclude package-path-1 package-path-2 ...]

Run go tests, optionally restricting which ones based on the only and exclude coommands.

EOF
}

function go_test() {
  local test_packages
  test_packages="$1"
  local tmpfile
  tmpfile=$(mktemp)
  # Interpolate tmpfile right now, so the trap set by the function
  # always work, even if ran outside the function body.
  # shellcheck disable=SC2064
  trap "rm \"$tmpfile\"" EXIT

  set +eo pipefail # so we still get the result if the test failed
  local test_exit_code
  # shellcheck disable=SC2086
  go test \
    -timeout 10m \
    -coverprofile=coverage.txt \
    -covermode=atomic \
    -race \
    -v \
    $test_packages | tee "$tmpfile"
  # Save the test exit code so we can return it after submitting the test run to the analytics.
  test_exit_code="${PIPESTATUS[0]}"
  set -eo pipefail # resume being strict about errors

  local xml
  xml=$(go-junit-report <"$tmpfile")
  # escape xml output properly for JSON
  local quoted_xml
  quoted_xml="$(echo "$xml" | jq -R -s '.')"

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
    --header "Authorization: Token token=\"$BUILDKITE_ANALYTICS_BACKEND_TEST_SUITE_API_KEY\";" \
    --header 'Content-Type: application/json' \
    --data-binary @-

  echo -e "\n--- :information_source: Succesfully uploaded test results to Buildkite analytics"

  return "$test_exit_code"
}

if [ "$1" == "-h" ]; then
  usage
  exit 1
fi

if [ -n "$1" ]; then
  FILTER_ACTION=$1
  shift
  FILTER_TARGETS=$*
fi

# Display to the user what kind of filtering is happening here
if [ -n "$FILTER_ACTION" ]; then
  echo -e "--- :information_source: \033[0;34mFiltering go tests: $FILTER_ACTION $FILTER_TARGETS\033[0m"
fi

# Buildkite analytics

# https://github.com/sourcegraph/sourcegraph/issues/28469
# TODO is that the best way to handle this?
go install github.com/jstemmer/go-junit-report@latest
asdf reshim golang

# TODO move to manifest
# https://github.com/sourcegraph/sourcegraph/issues/28469
BUILDKITE_ANALYTICS_BACKEND_TEST_SUITE_API_KEY=$(gcloud secrets versions access latest --secret="BUILDKITE_ANALYTICS_BACKEND_TEST_SUITE_API_KEY" --project="sourcegraph-ci" --quiet)

# For searcher
echo "--- comby install"
./dev/comby-install-or-upgrade.sh

# For code insights test
./dev/codeinsights-db.sh &
export CODEINSIGHTS_PGDATASOURCE=postgres://postgres:password@127.0.0.1:5435/postgres
export DB_STARTUP_TIMEOUT=360s # codeinsights-db needs more time to start in some instances.

# We have multiple go.mod files and go list doesn't recurse into them.
find . -name go.mod -exec dirname '{}' \; | while read -r d; do
  pushd "$d" >/dev/null

  # Separate out time for go mod from go test
  echo "--- $d go mod download"
  go mod download

  patterns="${FILTER_TARGETS[*]// /\\|}" # replace spaces with \| to have multiple patterns being matched
  case "$FILTER_ACTION" in
    exclude)
      TEST_PACKAGES=$(go list ./... | { grep -v "$patterns" || true; }) # -v to reject
      if [ -n "$TEST_PACKAGES" ]; then
        echo "--- $d go test"
        go_test "$TEST_PACKAGES"
      else
        echo "--- $d go test (skipping)"
      fi
      ;;
    only)
      TEST_PACKAGES=$(go list ./... | { grep "$patterns" || true; }) # select only what we need
      if [ -n "$TEST_PACKAGES" ]; then
        echo "--- $d go test"
        go_test "$TEST_PACKAGES"
      else
        echo "--- $d go test (skipping)"
      fi
      ;;
    *)
      TEST_PACKAGES="./..."
      echo "--- $d go test"
      go_test "$TEST_PACKAGES"
      ;;
  esac

  popd >/dev/null
done
