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

ALL_CMDS=$(echo {gitserver,indexer,query-runner,github-proxy,xlang-go,lsp-proxy,searcher,frontend,repo-updater,symbols})
if [ -n "$NOGORACED" ]; then
	echo "Go race detector disabled. You can enable it for specific commands by setting GORACED (e.g. GORACED=frontend,searcher or GORACED=all for all commands)"
	GORACED=''
elif [ -z "$GORACED" ]; then
	echo "Enabling GORACED=frontend by default."
	GORACED=frontend
elif [ "$GORACED" == "all" ]; then
	GORACED=$ALL_CMDS
fi

RACED_CMDS=''
if [ -n "$GORACED" ]; then
	IFS=', '
	for CMD in $GORACED; do
		echo "Go race detector enabled for: $CMD"
		RACED_CMD="sourcegraph.com/sourcegraph/sourcegraph/cmd/$CMD"
		if [ -z "$RACED_CMDS" ]; then
			RACED_CMDS=$RACED_CMD
		else
			RACED_CMDS="$RACED_CMDS $RACED_CMD"
		fi
	done
fi

NOT_RACED_CMDS=''
for CMD in $ALL_CMDS; do
	if [[ ! $RACED_CMDS =~ $CMD ]]; then
		RACED_CMD="sourcegraph.com/sourcegraph/sourcegraph/cmd/$CMD"
		if [ -z "$NOT_RACED_CMDS" ]; then
			NOT_RACED_CMDS=$RACED_CMD
		else
			NOT_RACED_CMDS="$NOT_RACED_CMDS $RACED_CMD"
		fi
	fi
done

if [ -n "$NOT_RACED_CMDS" ]; then
	echo $NOT_RACED_CMDS | xargs go install -v -gcflags="$GCFLAGS" -tags="$TAGS"
fi

if [ -n "$RACED_CMDS" ]; then
	echo $RACED_CMDS | xargs go install -v -gcflags="$GCFLAGS" -tags="$TAGS" -race
fi