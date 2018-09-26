#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/.bin"
    export PATH=$GOBIN:$PATH

    go install github.com/sourcegraph/sourcegraph/vendor/github.com/kevinburke/differ
    GO111MODULE=on differ bash -c 'go mod vendor && go mod tidy -v'
}

main "$@"
