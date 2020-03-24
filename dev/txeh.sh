#!/usr/bin/env bash

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

TXEH_PATH="${GOBIN}/txeh"

go build -o "${TXEH_PATH}" "github.com/txn2/txeh/util"

exec "${TXEH_PATH}" $@
