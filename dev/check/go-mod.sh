#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.."

    export GOBIN="$PWD/vendor/.bin"
    export PATH=$GOBIN:$PATH

    go install github.com/sourcegraph/sourcegraph/vendor/github.com/kevinburke/differ
    # Run any non-side-effecting command to ensure there's no Git diff present,
    # before running the "go mod" command below this one.
    #
    # We are seeing some nondeterministic build behavior and want to ensure the
    # problem is in "go mod" and not present before the tool even starts
    # running.
    #
    # https://buildkite.com/sourcegraph/sourcegraph/builds/20904#f00ca1a3-b30c-41fe-8896-f2e9e6473b61
    # https://github.com/sourcegraph/enterprise/issues/13416
    differ git fsck
    GO111MODULE=on differ bash -c 'go mod vendor && go mod tidy -v'
}

main "$@"
