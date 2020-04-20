#!/usr/bin/env bash

echo "--- lint dependencies"

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

if [ $# -eq 0 ]; then
  pkgs=('./...')
else
  pkgs=("$@")
fi

echo "--- go install"
go install -tags=dev -buildmode=archive "${pkgs[@]}"
asdf reshim

echo "--- lint"

# TODO when we add back golangci-lint use the same pattern we use for
# docsite.sh. We don't want to include its dependencies in our go.mod,
# especially since it is GPL code.

echo "go lint is disabled until we can get it to use less resources. https://github.com/sourcegraph/sourcegraph/issues/9193"
# Disable unused since it uses too much CPU/mem
#golangci-lint run -e unused ${pkgs}
