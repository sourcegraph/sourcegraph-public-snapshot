#!/usr/bin/env bash

set -e

# For searcher
echo "--- :arrow_down: comby install"
(
  set -x
  ./dev/comby-install-or-upgrade.sh
)


echo "--- :arrow_right: Running CodeInsightsDB"
(
  set -x
  # For code insights test
  ./dev/codeinsights-db.sh &
  export CODEINSIGHTS_PGDATASOURCE=postgres://postgres:password@127.0.0.1:5435/postgres
  export DB_STARTUP_TIMEOUT=120s # codeinsights-db needs more time to start in some instances.
)

echo "--- :go: Getting richgo"
(
  set -x
  wget https://github.com/kyoh86/richgo/releases/download/v0.3.9/richgo_0.3.9_linux_amd64.tar.gz
  tar zxf richgo_0.3.9_linux_amd64.tar.gz
)

# We have multiple go.mod files and go list doesn't recurse into them.
find . -name go.mod -exec dirname '{}' \; | while read -r d; do
  pushd "$d" >/dev/null

  # Separate out time for go mod from go test
  echo "--- :go: $d go mod download"
  go mod download

  echo "--- :go: $d go test"
  ./richgo_0.3.9_linux_amd64/richgo test -timeout 10m -coverprofile=coverage.txt -covermode=atomic -race ./...

  popd >/dev/null
done
