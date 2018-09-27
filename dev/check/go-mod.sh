#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/.bin"
    export PATH=$GOBIN:$PATH

    ./dev/go-mod-update.sh
    if ! git diff-index --quiet HEAD --; then
        echo "FAIL: working directory has changed"
        git diff
        git status
        exit 2
    fi
}

main "$@"
