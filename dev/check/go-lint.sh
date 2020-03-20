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

golangci-lint run ${pkgs}
