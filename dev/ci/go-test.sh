#!/usr/bin/env bash

set -euo pipefail

function usage {
  cat <<EOF
Usage: go-test.sh [only|exclude package-path-1 package-path-2 ...]

Run go tests, optionally restricting which ones based on the only and exclude coommands.

EOF
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
        # shellcheck disable=SC2086
        go test -timeout 10m -coverprofile=coverage.txt -covermode=atomic -race $TEST_PACKAGES
      else
        echo "--- $d go test (skipping)"
      fi
      ;;
    only)
      TEST_PACKAGES=$(go list ./... | { grep "$patterns" || true; }) # select only what we need
      if [ -n "$TEST_PACKAGES" ]; then
        echo "--- $d go test"
        # shellcheck disable=SC2086
        go test -timeout 10m -coverprofile=coverage.txt -covermode=atomic -race $TEST_PACKAGES
      else
        echo "--- $d go test (skipping)"
      fi
      ;;
    *)
      TEST_PACKAGES="./..."
      echo "--- $d go test"
      # shellcheck disable=SC2086
      go test -timeout 10m -coverprofile=coverage.txt -covermode=atomic -race $TEST_PACKAGES
      ;;
  esac

  popd >/dev/null
done
