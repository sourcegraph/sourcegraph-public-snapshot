#!/bin/bash
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

generate_graphql=false
generate_schema=false
rebuild_symbols=false
cmdlist=""
all_cmds=false
failed=false

for i; do
	case $i in
	"cmd/frontend/graphqlbackend/schema.graphql")
		generate_graphql=true
		;;
	schema/*.json)
		generate_schema=true
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

	case $i in
	cmd/symbols/*)
		rebuild_symbols=true
		;;
    esac
done

$generate_graphql && { go generate github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend || failed=true; }
$generate_schema && { go generate github.com/sourcegraph/sourcegraph/schema || failed=true; }
$rebuild_symbols && { rm -rf /tmp/symbols-cache; env IMAGE=dev-symbols cmd/symbols/build.sh; }
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
