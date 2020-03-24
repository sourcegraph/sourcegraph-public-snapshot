#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

echo "--- go install txeh"
go install github.com/txn2/txeh

if [ -x "$(command -v asdf)" ]; then
    asdf reshim golang
fi

SOURCEGRAPH_HTTPS_DOMAIN="${SOURCEGRAPH_HTTPS_DOMAIN:-"sourcegraph.test"}"

echo "--- adding ${SOURCEGRAPH_HTTPS_DOMAIN} to '/etc/hosts' (you may need to enter your password)"

sudo txeh add 127.0.0.1 "${SOURCEGRAPH_HTTPS_DOMAIN}"

echo "--- printing '/etc/hosts'"

txeh show
