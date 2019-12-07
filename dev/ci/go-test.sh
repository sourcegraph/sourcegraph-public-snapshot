#!/usr/bin/env bash

set -e

# For symbol tests
echo "--- build libsqlite"
./cmd/symbols/build.sh buildLibsqlite3Pcre

# For searcher and replacer tests
echo "--- comby install"
./dev/comby-install-or-upgrade.sh

# Seperate out time for go mod from go test
echo "--- go mod download"
go mod download

echo "--- go test"
go test -timeout 4m -coverprofile=coverage.txt -covermode=atomic -race ./...
