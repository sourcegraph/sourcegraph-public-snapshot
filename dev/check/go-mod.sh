#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/.bin"
    export PATH=$GOBIN:$PATH

    ./dev/go-mod-update.sh
    if [ ! -z "$(git status --porcelain)" ]; then
        echo "FAIL: working directory has changed"
        git diff
        git status
        exit 2
    fi
}

main "$@"
