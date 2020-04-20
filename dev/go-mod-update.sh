#!/usr/bin/env bash

set -euxf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

export GOPRIVATE=github.com/mattn/go-sqlite3

go get -u all
go mod tidy
go test -short -failfast ./... >/dev/null
