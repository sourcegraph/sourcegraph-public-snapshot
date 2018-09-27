#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/.."

    export GO111MODULE=on
    # We cannot drop the ../sourcegraph module replacement on CI yet because it
    # tries to clone the repo over HTTPS and prompts for a password.
    [ -z "$BUILDKITE" ] && go mod edit -dropreplace github.com/sourcegraph/sourcegraph
    go mod vendor
    go mod tidy -v
    [ -z "$BUILDKITE" ] && go mod edit -replace github.com/sourcegraph/sourcegraph=../sourcegraph
    go mod tidy # TODO(slimsag): I don't understand why this is needed, but it is.
}

main "$@"
