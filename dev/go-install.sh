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

go install -v -gcflags="$GCFLAGS" -tags="$TAGS" sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,query-runner,github-proxy,xlang-go,lsp-proxy,searcher,frontend,repo-updater,symbols}

if [ -n "$NOGORACED" ]; then
	echo "Go race detector disabled. You can enable it for specific commands by setting GORACED (e.g. GORACED=frontend,searcher or GORACED=... for all commands)"
	GORACED=''
elif [ -z "$GORACED" ]; then
	echo "Enabling GORACED=frontend by default."
	GORACED=frontend
fi

if [ -n "$GORACED" ]; then
	IFS=','
	for CMD in $GORACED; do
		echo "Go race detector enabled for: $CMD"
		go install -race -v -gcflags="$GCFLAGS" -tags="$TAGS" sourcegraph.com/sourcegraph/sourcegraph/cmd/$CMD
	done
fi
