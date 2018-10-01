#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir
OSS_GO_DIRS="$(../sourcegraph/dev/watchdirs.sh | sed 's|[^ ]* *|../sourcegraph/&|g')"
GO_DIRS="$OSS_GO_DIRS cmd dev"

dirs_starstar() {
	for i; do echo "'$i/**/*.go'"; done
}

dirs_path() {
	for i; do echo "-path $i"; done
}

useChokidar() {
	echo >&2 "Using chokidar."
	# eval so the expansion can produce quoted things, and eval can eat the
	# quotes, so it doesn't try to expand wildcards.
	eval exec chokidar --silent $(dirs_starstar $GO_DIRS) ../sourcegraph/cmd/frontend/graphqlbackend/schema.graphql "'../sourcegraph/schema/*.json'" -c "'./dev/handle-change.sh {path}'"
}

useInotifywrapper() {
	echo >&2 "Using inotifywrapper."
	exec dev/inotifywrapper/inotifywrapper $(dirs_path $GO_DIRS) \
		-match '\.go$' \
		-match '../sourcegraph/cmd/frontend/graphqlbackend/schema\.graphql' \
		-match '../sourcegraph/schema/.*.json' \
		-cmd './dev/handle-change.sh'
}

case $(which inotifywait 2>/dev/null) in
"")	useChokidar
	;;
*)	if ( cd dev/inotifywrapper; go build ); then
		useInotifywrapper
	else
		echo >&2 "Can't build inotifywrapper: Falling back on chokidar."
		useChokidar
	fi
	;;
esac
