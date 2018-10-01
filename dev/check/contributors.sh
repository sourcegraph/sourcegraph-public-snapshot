#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/vendor/.bin"
    export PATH=$GOBIN:$PATH
    go install github.com/sourcegraph/sourcegraph/vendor/github.com/kevinburke/differ

    differ ./dev/contributors.sh
}

main "$@"
