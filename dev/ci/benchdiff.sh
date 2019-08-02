#!/usr/bin/env bash

set -euo pipefail

old="${1:-""}"
new="${2:-""}"
save=${BENCHSAVE:-"0"}

function usage() {
  echo "Usage: ./benchdiff.sh <old.txt> <new.txt>"
}

if [ ! -f "$old" ] || [ ! -f "$new" ]; then
  usage
  exit 1
fi

deltas=$(benchstat -csv -split "" "$old" "$new" | xsv select 6 | grep -v 'delta|~')
changes=$(echo "$deltas" | wc -l)

if [ "$save" -ne 0 ] && [ "$changes" -ne 0 ]; then
  benchsave "$old" "$new"
else
  benchstat "$old" "$new"
fi

regressions=$(echo "$deltas" | grep -c '^+')
if [ "$regressions" -ne 0 ]; then
  echo "WARNING: Performance regressions detected. Please investigate."
  exit 1
fi
