#!/usr/bin/env bash

set -e

# For symbol tests
echo "--- build libsqlite"
./dev/libsqlite3-pcre/build.sh

# For searcher
echo "--- comby install"
./dev/comby-install-or-upgrade.sh

# For code insights test
./dev/codeinsights-db.sh &
export CODEINSIGHTS_PGDATASOURCE=postgres://postgres:password@127.0.0.1:5435/postgres
export DB_STARTUP_TIMEOUT=120s # codeinsights-db needs more time to start in some instances.

# We have multiple go.mod files and go list doesn't recurse into them.
find . -name go.mod -exec dirname '{}' \; | while read -r d; do
  pushd "$d" >/dev/null

  # Separate out time for go mod from go test
  echo "--- $d go mod download"
  go mod download

  echo "--- $d go test"
  go test -timeout 4m -coverprofile=coverage.txt -covermode=atomic -race ./...

  popd >/dev/null
done
