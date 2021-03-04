#!/bin/bash
set -euo pipefail

if [[ $# -le 0 ]]; then
  echo 'usage: ./combine.sh file1.lsif file2.lsif file3.lsif ...'
  exit 1
fi

BASE_FILE=$1

# get the project root of the base file for the merge
PROJECT_ROOT=$(head -n 1 "$BASE_FILE" | jq -r '.projectRoot')

# IDs are in the range 1-(num lines), so we need to increment all the IDs in the other dumps
NUM_IDS=$(wc -l "$BASE_FILE" | awk '{print $1}')
FILES=("$BASE_FILE")
shift

while [[ $# -gt 0 ]]; do
  NEXT_FILE=$1
  NEXT_ROOT=$(head -n 1 "$BASE_FILE" | jq -r '.projectRoot')
  ESCAPED_PROJECT_ROOT=$(printf '%s\n' "$PROJECT_ROOT" | sed 's:[\\/&]:\\&:g;$!s/$/\\/')
  ESCAPED_NEXT_ROOT=$(printf '%s\n' "$NEXT_ROOT" | sed 's:[][\\/.^$*]:\\&:g')

  # tail: skip the meta and project vertexes
  # jq: increment all IDs, and all references to IDs from edges.
  # EXCEPT, don't change the ID on the contains edge which maps project vertex to document,
  # since that's always vertex #2 and we only include it once
  # sed: replace whatever project root the dump we're currently merging with
  # the project root from the base dump
  # output: to a temporary file in a subprocess, so we can run all the jqs in parallel.
  # jq is by far the bottlneck but anything else that's safe would be really convoluted.
  tail -n +3 "$NEXT_FILE" | jq -rc "(.id,.inV,.outV,.document,.inVs[]? | select(. != null)) |= . + $NUM_IDS" | sed "s/$ESCAPED_NEXT_ROOT/$ESCAPED_PROJECT_ROOT/" >"/tmp/$NEXT_FILE.tmp" &
  FILES+=("/tmp/$NEXT_FILE.tmp")

  NUM_NEW_IDS=$(wc -l "$NEXT_FILE" | awk '{print $1}')
  NUM_IDS=$((NUM_IDS + NUM_NEW_IDS - 2))
  shift
done

wait
for FILE in "${FILES[@]}"; do
  cat "$FILE"
done
