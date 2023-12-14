#!/usr/bin/env bash

ARGS=("$@")

if [[ ${BUILDKITE_RETRY_COUNT:-0} == 1 ]]; then
  ARGS=("${ARGS[@]}" "--force-rebuild")
fi

pnpm chromatic ${ARGS[*]}
