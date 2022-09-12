#!/usr/bin/env bash

set -e

# Use .bin outside of schema since schema dir is watched by watchman.
export GOBIN="$PWD/../.bin"
export GO111MODULE=on

go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler@v0.0.0-20211011105148-2e30f7bacbe1

# shellcheck disable=SC2010
schemas="$(ls -- *.schema.json | grep -v json-schema-draft)"

# shellcheck disable=SC2086
"$GOBIN"/go-jsonschema-compiler -o schema.go -pkg schema $schemas

gofmt -s -w ./*.go
