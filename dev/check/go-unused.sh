#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/vendor/.bin"
export PATH=$GOBIN:$PATH

go install github.com/sourcegraph/sourcegraph/vendor/honnef.co/go/tools/cmd/unused

# To allow unused Go identifiers, add an entry below of the form `file:U1000`. (U1000 is the "ID" of
# the unused check in the honnef.co/go/tools suite.) It is only possible to ignore entire files, not
# individual identifiers.
#
# TODO: Remove these exceptions. See https://github.com/sourcegraph/sourcegraph/issues/11799.
IGNORE_UNUSED=(
    'github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend/git_ref.go:U1000'
    'github.com/sourcegraph/sourcegraph/pkg/endpoint/endpoint.go:U1000'
    'github.com/sourcegraph/sourcegraph/xlang/gobuildserver/build_server.go:U1000'
    'github.com/sourcegraph/sourcegraph/xlang/proxy/log.go:U1000'
    'github.com/sourcegraph/sourcegraph/xlang/proxy/server_proxy.go:U1000'
)

function join_by { local IFS="$1"; shift; echo "$*"; }

echo '(go) unused...'
unused -ignore "$(join_by " " "${IGNORE_UNUSED[@]}")" ./...
