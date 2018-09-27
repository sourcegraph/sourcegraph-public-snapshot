#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/.."

    export GO111MODULE=on
    go mod edit -dropreplace github.com/sourcegraph/sourcegraph
    go mod vendor
    go mod tidy -v
    go mod edit -replace github.com/sourcegraph/sourcegraph=../sourcegraph
    go mod tidy # TODO(slimsag): I don't understand why this is needed, but it is.
}

main "$@"
