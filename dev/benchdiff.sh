#!/usr/bin/env bash

set -euo pipefail

function usage() {
    echo "Usage:   ./benchdiff.sh <old> <new> <path> <iter>"
    echo "Example: ./benchdiff.sh master^1 master ./... 10"
}

old=${1:-$(git rev-parse HEAD^)}
new=${2:-$(git rev-parse HEAD)}
path=${3:-"./..."}
iter=${4:-"10"}
save=${BENCHSAVE:-"0"}

git checkout "$old"
oldout="$old.old.bench.txt"
go test -run=^$ -bench=. -benchmem -count="$iter" "$path" | tee "$oldout"

git checkout "$new"
newout="$new.new.bench.txt"
go test -run=^$ -bench=. -benchmem -count="$iter" "$path" | tee "$newout"

deltas=$(benchstat -csv "$oldout" "$newout" | xsv select 6 | grep -v 'delta|~')
changes=$(echo "$deltas" | wc -l)

if [ "$save" -ne 0 ] && [ "$changes" -ne 0 ]; then
  benchsave "$oldout" "$newout" | tee "$old-$new.bench.txt"
else
  benchstat "$oldout" "$newout" | tee "$old-$new.bench.txt"
fi

regressions=$(echo "$deltas" | grep -c '^+')
exit "$regressions"
