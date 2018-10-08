#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir
OSS_GO_DIRS="$(../sourcegraph/dev/watchdirs.sh | sed 's|[^ ]* *|../sourcegraph/&|g')"
GO_DIRS="$OSS_GO_DIRS cmd dev"

dirs_starstar() {
	for i; do echo "'$i/**/*.go'"; done
}

# eval so the expansion can produce quoted things, and eval can eat the
# quotes, so it doesn't try to expand wildcards.
eval exec chokidar --silent $(dirs_starstar $GO_DIRS) ../sourcegraph/cmd/frontend/graphqlbackend/schema.graphql "'../sourcegraph/schema/*.json'" -c "'./dev/handle-change.sh {path}'"
