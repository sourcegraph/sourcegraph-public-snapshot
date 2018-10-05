#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/vendor/.bin"
    export PATH=$GOBIN:$PATH

    GO111MODULE=on go mod vendor
    git diff --exit-code -- . ':!go.sum'
}

main "$@"
