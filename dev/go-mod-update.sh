#!/usr/bin/env bash

set -euxf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

export GOPRIVATE=github.com/mattn/go-sqlite3

go get -u all

# Pins

## Newer versions seems to be enforcing something to do with resource
## names. This is causing our use of our k8s client to panic.
go get github.com/golang/protobuf@v1.3.5

# Cleanup and validate everything still works
go mod tidy
go test -short -failfast ./... >/dev/null
