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

case $1 in
cmd/*)	cmd=${1#cmd/}
			cmd=${cmd%%/*}
			rebuilt=$(./dev/go-install.sh -v $cmd)
			;;
*)			rebuilt=$(./dev/go-install.sh -v)
			;;
esac

echo >&2 "Rebuilt: $rebuilt"
[ -n "$rebuilt" ] && $GOREMAN run restart $rebuilt
