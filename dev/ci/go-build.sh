#!/usr/bin/env bash

set -e

# Seperate out time for go mod from go install
echo "--- go mod download"
go mod download

echo "--- go generate"
go generate ./...

echo "--- go install"
go install -tags dist ./cmd/... ./enterprise/cmd/...
