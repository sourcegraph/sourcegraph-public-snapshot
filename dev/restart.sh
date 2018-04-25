#!/bin/bash
set -e

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

./dev/go-install.sh

cmd=$(echo $1 | sed -E 's/cmd\/([^/]+)\/.*/\1/g')
if [ "$cmd" == "$1" ]; then
    # Changed file was not in a cmd subdirectory, so we need to pessimistically restart everything.
    $GOREMAN run restart gitserver indexer query-runner github-proxy xlang-go lsp-proxy searcher frontend repo-updater symbols
else
    $GOREMAN run restart $cmd
fi
