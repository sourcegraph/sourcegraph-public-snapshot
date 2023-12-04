#!/usr/bin/env bash

set -ex
MAX_DURATION_IN_SECONDS=600 # seconds

tmp_log=$(mktemp)
trap 'rm -f ${tmp_log}' EXIT

while : ; do
  if [ "$SECONDS" -gt "$MAX_DURATION_IN_SECONDS" ]; then
    echo "--- ✂️ timeout reached, aborting".
    exit 1
  fi
  if pnpm install --frozen-lockfile --fetch-timeout 30000 --silent "$@" 2> >(tee "$tmp_log">&2); then
    break
  fi

  if grep -q "An unexpected error occurred" < "$tmp_log"; then
    echo "--- 🔌 possible transient error found, trying again..."
  else
    break
  fi
done
