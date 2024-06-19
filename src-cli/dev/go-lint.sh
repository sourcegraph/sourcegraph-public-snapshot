#!/usr/bin/env bash

trap "echo ^^^ +++" ERR

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/.."

export GOBIN="$PWD/dev/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

if [ $# -eq 0 ]; then
  pkgs=('./...')
else
  pkgs=("$@")
fi

"./dev/golangci-lint.sh" --config .golangci.yml run "${pkgs[@]}"
