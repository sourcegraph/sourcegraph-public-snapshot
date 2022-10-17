#!/usr/bin/env bash

MAX_DURATION_IN_SECONDS=600 # seconds

tmp_log=$(mktemp)
trap 'rm -f ${tmp_log}' EXIT

while : ; do
  if [ "$SECONDS" -gt "$MAX_DURATION_IN_SECONDS" ]; then
    echo "--- ✂️ timeout reached, aborting".
    exit 1
  fi
  if yarn --immutable --network-timeout 30000 "$@" 2> >(tee "$tmp_log">&2); then
    break
  fi

  if grep -q "An unexpected error occurred" < "$tmp_log"; then
    echo "--- 🔌 possible transient error found, trying again..."
  else
    break
  fi
done
