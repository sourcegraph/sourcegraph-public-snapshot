#!/usr/bin/env bash

set -euxf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

export GOPRIVATE=github.com/mattn/go-sqlite3

go get -u all

# Pins

## Newer versions seems to be enforcing something to do with resource
## names. This is causing our use of our k8s client to panic.
go get github.com/golang/protobuf@v1.3.5

## Newer versions removed some types in the endpoint package we relied on. Unsure how to fix yet, so punting.
go get github.com/aws/aws-sdk-go-v2@v0.20.0

# Cleanup and validate everything still works
go mod tidy
go test -short -failfast ./... >/dev/null
