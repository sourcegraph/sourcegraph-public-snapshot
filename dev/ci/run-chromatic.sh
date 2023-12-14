#!/usr/bin/env bash

ARGS=("$@")

if [[ ${BUILDKITE_RETRY_COUNT:-0} == 1 ]]; then
  # Chromatic fails with exit-code 1 if the commit stays the same and instructs one to add `--force-rebuild`
  # So we detect when a build is retried and then add the flag accordingly
  ARGS=("${ARGS[@]}" "--force-rebuild")
fi

pnpm chromatic "${ARGS[*]}"
