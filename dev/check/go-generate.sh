#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/vendor/.bin"
    export PATH=$GOBIN:$PATH

    go install github.com/sourcegraph/sourcegraph/vendor/golang.org/x/tools/cmd/stringer
    go install github.com/sourcegraph/sourcegraph/vendor/github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler
    go install github.com/sourcegraph/sourcegraph/vendor/github.com/kevinburke/differ

    # Runs generate.sh and ensures no files changed. This relies on the go
    # generation that ran are idempotent.

    differ ./dev/generate.sh
}

main "$@"
