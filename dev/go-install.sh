#!/bin/bash

mkdir -p .bin
export GOBIN=$PWD/.bin

go get sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/mattn/goreman

TAGS='all=dev'
if [ -n "$DELVE" ]; then
	echo 'Building with optimizations disabled (for debugging). Make sure you have at least go1.10 installed.'
	GCFLAGS='all=-N -l'
	TAGS="$TAGS delve"
fi

go install -race -v -gcflags="$GCFLAGS" -tags="$TAGS" sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,query-runner,github-proxy,xlang-go,lsp-proxy,searcher,frontend,repo-updater,symbols}
