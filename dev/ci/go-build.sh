#!/usr/bin/env bash

set -e

# Separate out time for go mod from go install
echo "--- go mod download"
go mod download

echo "--- go generate"
export GOBIN="${PWD}/.bin"
go install golang.org/x/tools/cmd/goimports
go generate ./...

echo "--- go install"
go install -tags dist ./cmd/... ./enterprise/cmd/...
