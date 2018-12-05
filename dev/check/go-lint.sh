#!/bin/bash
set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

pkgs=${@:-./...}

go install github.com/golangci/golangci-lint/cmd/golangci-lint

echo "--- go install"
go install -buildmode=archive ${pkgs}

echo "--- lint"
git fetch origin master
rev="${BUILDKITE_PULL_REQUEST_BASE_BRANCH:-HEAD~}"
golangci-lint run -v ${pkgs} --new-from-rev ${rev}
