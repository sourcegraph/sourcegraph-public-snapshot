#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir
GO_DIRS="$(./dev/watchdirs.sh) ${WATCH_ADDITIONAL_GO_DIRS}"

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
    eval exec chokidar --silent \
        $(dirs_starstar $GO_DIRS) \
        cmd/frontend/graphqlbackend/schema.graphql \
        "'schema/*.json'" \
        "'cmd/symbols/**/*'" \
        "'cmd/symbols/.ctags.d/*'" \
        -c "'./dev/handle-change.sh {path}'"
}

execInotifywrapper() {
    echo >&2 "Using inotifywrapper."
    set -e
    pushd dev/inotifywrapper
    go build
    popd
    exec dev/inotifywrapper/inotifywrapper $(dirs_path $GO_DIRS) \
        -match '\.go$' \
        -match 'cmd/frontend/graphqlbackend/schema\.graphql' \
        -match 'schema/.*.json' \
        -cmd './dev/handle-change.sh'
}

execWatchman() {
    echo >&2 "Using watchman."
    set -e
    pushd dev/watchmanwrapper
    go build
    popd
    exec dev/watchmanwrapper/watchmanwrapper dev/handle-change.sh <<-EOT
["subscribe", ".", "gochangewatch", {
  "expression": ["anyof",
    ["suffix", "go"],
    ["dirname", "cmd/symbols"],
    ["dirname", "schema"],
    ["name", "cmd/frontend/graphqlbackend/schema.graphql", "wholename"]
  ],
  "fields": ["name"]
}]
EOT
}

[ -x "$(command -v watchman)" ] && execWatchman
[ -x "$(command -v inotifywait)" ] && execInotifywrapper

useChokidar
