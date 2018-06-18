#!/bin/bash
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

useChokidar() {
	echo >&2 "Using chokidar."
	exec chokidar --silent "cmd/**/*.go" "pkg/**/*.go" "vendor/**/*.go" "cmd/frontend/internal/graphqlbackend/schema.graphql" "schema/*.json" -c "./dev/handle-change.sh {path}"
}

useInotifywrapper() {
	echo >&2 "Using inotifywrapper."
	exec dev/inotifywrapper/inotifywrapper -path vendor -path pkg -path cmd \
		-match '\.go$' \
		-match 'cmd/frontend/internal/graphqlbackend/schema\.graphql' \
		-match 'schema/.*.json' \
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
