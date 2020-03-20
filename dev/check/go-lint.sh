#!/bin/bash

echo "--- lint dependencies"

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

pkgs=${@:-./...}

go install github.com/golangci/golangci-lint/cmd/golangci-lint

echo "--- go install"
go install -tags=dev -buildmode=archive ${pkgs}
asdf reshim

echo "--- lint"

echo "go lint is disabled until we can get it to use less resources. https://github.com/sourcegraph/sourcegraph/issues/9193"
# Disable unused since it uses too much CPU/mem
#golangci-lint run -e unused ${pkgs}
