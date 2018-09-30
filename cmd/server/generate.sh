#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -ex

rm -f dockerfile.go
cp "$(go list -f '{{.Dir}}' github.com/sourcegraph/sourcegraph/cmd/server)/dockerfile.go" ./dockerfile.go
