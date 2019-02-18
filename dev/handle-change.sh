#!/bin/bash
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

generate_graphql=false
generate_schema=false
cmdlist=""
all_cmds=false
failed=false
onlyBuildDevSymbols=false

for i; do
	case $i in
	"cmd/frontend/graphqlbackend/schema.graphql")
		generate_graphql=true
		;;
	schema/*.json)
		generate_schema=true
		;;
    cmd/symbols/.ctags.d/*)
        onlyBuildDevSymbols=true
        ;;
    cmd/symbols/Dockerfile)
        onlyBuildDevSymbols=true
        ;;
	cmd/*)
		cmd=${i#cmd/}
		cmd=${cmd%%/*}
		case " $cmdlist " in
		" $cmd ")
			;;
		*)
			cmdlist="$cmdlist $cmd"
			;;
		esac
		;;
	*)
		all_cmds=true
		;;
	esac
done

$generate_graphql && { go generate github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend || failed=true; }
$generate_schema && { go generate github.com/sourcegraph/sourcegraph/schema || failed=true; }
$onlyBuildDevSymbols && {
    set -e
    ./dev/ts-script cmd/symbols/build.ts buildDockerImage --dockerImageName dev-symbols
    [ -n "$GOREMAN" ] && $GOREMAN run restart symbols
    exit
}

if $all_cmds; then
	rebuilt=$(./dev/go-install.sh -v | tr '\012' ' ')
	[ $? == 0 ] || failed=true
elif [ -n "$cmdlist" ]; then
	rebuilt=$(./dev/go-install.sh -v $cmdlist | tr '\012' ' ')
	[ $? == 0 ] || failed=true
fi

if [ -n "$rebuilt" ]; then
	echo >&2 "Rebuilt: $rebuilt"
	[ -n "$rebuilt" ] && [ -n "$GOREMAN" ] && $GOREMAN run restart $rebuilt
else
	echo >&2 "Nothing to rebuild or rebuilds failed."
fi

! $failed
