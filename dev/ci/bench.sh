#!/usr/bin/env bash

set -euo pipefail

commit=$(git rev-parse "${1:-"HEAD"}")
iter=${2:-"10"}
path=${3:-"./..."}

echo "---- Running benchmarks @$commit ----"
git checkout "$commit"
go test -run=^$ -bench=. -benchmem -count="$iter" "$path"
