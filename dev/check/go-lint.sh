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
"./dev/golangci-lint.sh" --config .golangci.enforced.yml run "${pkgs[@]}"
