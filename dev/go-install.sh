#!/bin/bash
# build commands, optionally with or without race detector.
# a list of every command we know about, to use by default
all_commands=" gitserver indexer query-runner github-proxy xlang-go lsp-proxy searcher frontend repo-updater symbols "

# handle options
verbose=false
while getopts 'v' o; do
	case $o in
	v)	verbose=true;;
	\?)	echo >&2 "usage: go-install.sh [-v] [commands]"
		exit 1
		;;
	esac
done
shift $(expr $OPTIND - 1)

# check provided commands
ok=true
case $# in
0)	commands=$all_commands;;
*)	commands=" $* "
	for cmd in $commands; do
		case $all_commands in
		*" $cmd "*)	;;
		*)	echo >&2 "unknown command: $cmd"
			ok=false
			;;
		esac
	done
	;;
esac

$ok || exit 1

mkdir -p .bin
export GOBIN=$PWD/.bin

if ! go install \
	github.com/sourcegraph/sourcegraph/vendor/github.com/mattn/goreman \
	github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-archive-index \
	github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
	github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-webserver; then
	echo >&2 "failed to install prerequisites, aborting."
	exit 1
fi

TAGS='dev'
if [ -n "$DELVE" ]; then
	echo >&2 'Building with optimizations disabled (for debugging). Make sure you have at least go1.10 installed.'
	GCFLAGS='all=-N -l'
	TAGS="$TAGS delve"
fi

# build a list of "cmd,true" and "cmd,false" pairs to indicate whether each command
# wants its own flags. we can't use variable names with the command in them because
# some commands have hyphens.
cmdlist=""
anyraced=false
case $GORACED in
all)	for cmd in $commands; do
		cmdlist="$cmdlist $cmd,true"
		anyraced=true
	done
	;;
*)	for cmd in $commands; do
		case " $GORACED " in
		*" $cmd "*)
			raced=true
			anyraced=true
			;;
		*)
			raced=false
			;;
		esac
		cmdlist="$cmdlist $cmd,$raced"
	done
	;;
esac

if ! $anyraced; then
	echo >&2 "Go race detector disabled. You can enable it for specific commands by setting GORACED (e.g. GORACED=frontend,searcher or GORACED=all for all commands)"
fi

failed=""
for cmd in $cmdlist; do
	raced=${cmd##*,}
	cmd=${cmd%,*}
	if go install -v -gcflags="$GCFLAGS" -tags="$TAGS" -race=$raced github.com/sourcegraph/sourcegraph/cmd/$cmd >&2; then
		$verbose && echo "$cmd"
	else
		failed="$failed $cmd"
	fi
done

if [ -n "$failed" ]; then
	echo >&2 "failed to build:$failed"
	exit 1
fi
