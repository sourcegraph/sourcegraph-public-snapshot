#!/bin/bash
set -e

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

if [ "$1" == "cmd/frontend/internal/graphqlbackend/schema.graphql" ]; then
    go generate github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend
    exit
fi

if [[ $1 =~ schema/.*\.json ]]; then
    go generate github.com/sourcegraph/sourcegraph/schema
    exit
fi

./dev/go-install.sh

cmd=$(echo $1 | sed -E 's/cmd\/([^/]+)\/.*/\1/g')
if [ "$cmd" == "$1" ]; then
    # Changed file was not in a cmd subdirectory, so we need to pessimistically restart everything.
    $GOREMAN run restart gitserver indexer query-runner github-proxy xlang-go lsp-proxy searcher frontend repo-updater symbols
else
    $GOREMAN run restart $cmd
fi
