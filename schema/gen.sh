#!/usr/bin/env bash

set -e

# Use .bin outside of schema since schema dir is watched by watchman.
export GOBIN="$PWD/../.bin"
export GO111MODULE=on

go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler

# shellcheck disable=SC2010
schemas="$(ls -- *.schema.json | grep -v json-schema-draft)"

# shellcheck disable=SC2086
"$GOBIN"/go-jsonschema-compiler -o schema.go -pkg schema $schemas

stringdata() {
  # shellcheck disable=SC2039
  target="${1/.schema.json/_stringdata.go}"
  "$GOBIN"/stringdata -i "$1" -name "$2" -pkg schema -o "$target"
}

gofmt -s -w ./*.go
