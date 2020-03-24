#!/usr/bin/env bash

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

CADDY_PATH="${GOBIN}/caddy"

go build -o "${CADDY_PATH}" "github.com/caddyserver/caddy/v2/cmd/caddy"

exec "${CADDY_PATH}" $@
