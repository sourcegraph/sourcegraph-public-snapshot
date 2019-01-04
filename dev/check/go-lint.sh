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

echo "--- lint"
if [ -n "$BUILDKITE_PULL_REQUEST_BASE_BRANCH" ]; then
    git fetch origin ${BUILDKITE_PULL_REQUEST_BASE_BRANCH}
    base="origin/${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
else
    git fetch origin master
    base="HEAD~"
fi

rev=$(git merge-base ${base} HEAD)
golangci-lint run --build-tags=dev -v ${pkgs} --new-from-rev ${rev} --deadline 5m
