#!/usr/bin/env bash

set -e

pushd ./dev/sg >/dev/null

# Separate out time for go mod from go test
echo "--- $d go mod download"
go mod download

echo "--- $d go test"
go test -timeout 5m -coverprofile=coverage.txt -covermode=atomic -race ./...

popd >/dev/null
