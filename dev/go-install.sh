#!/bin/bash

mkdir -p .bin
export GOBIN=$PWD/.bin

go get sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/mattn/goreman

TAGS='dev'
if [ -n "$DELVE" ]; then
	echo 'Building with optimizations disabled (for debugging)'
	GCFLAGS='-N -l'
	TAGS="$TAGS delve"
fi

SAVED_GCFLAGS="$PWD/.bin/GCFLAGS"
touch $SAVED_GCFLAGS

if [ "$GCFLAGS" != "$(cat $SAVED_GCFLAGS)" ]; then
	# Until go 1.10 is released go install does not
	# rebuild the requested packages and dependencies
	# if there are cached versions even if gcflags have changed.
	# see https://github.com/golang/go/issues/19340
	# To workaround this, nuke cached packages that we might care to debug
	# if we detect that GCFLAGS has changed.
	echo 'Clearing cached packages because debug state changed (build may take a while)'
	rm -rf $GOPATH/pkg/*/sourcegraph.com/sourcegraph/sourcegraph/*
	echo $GCFLAGS > $SAVED_GCFLAGS
fi

go install -v -gcflags="$GCFLAGS" -tags="$TAGS" sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,github-proxy,xlang-go,lsp-proxy,searcher,frontend,repo-list-updater}
