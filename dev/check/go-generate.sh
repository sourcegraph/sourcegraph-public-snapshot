#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    mkdir -p ui/assets

    export GOBIN="$PWD/.bin"
    export PATH=$GOBIN:$PATH

    go install golang.org/x/tools/cmd/stringer
    go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler

    # Runs generate.sh and ensures no files changed. This relies on the go
    # generation that ran are idempotent.

    ./dev/generate.sh
    git diff --exit-code -- . ':!go.sum'
}

main "$@"
