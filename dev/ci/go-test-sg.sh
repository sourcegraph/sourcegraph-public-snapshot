#!/usr/bin/env bash

set -e

pushd ./dev/sg >/dev/null

# Separate out time for go mod from go test
echo "--- go mod download"
go mod download

echo "--- go test"
go test -timeout 5m -coverprofile=coverage.txt -covermode=atomic -race ./...

popd >/dev/null
