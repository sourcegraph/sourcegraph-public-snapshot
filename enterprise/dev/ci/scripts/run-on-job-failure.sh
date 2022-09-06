#!/usr/bin/env bash

outcome=$(buildkite-agent step get "outcome" --step "$1")

echo "Getting outcome for $1: $outcome"

if [ "$outcome" == "hard_failure" ]; then
  echo "run-on-job-failure.sh: $1 step failed with $outcome"
else
  echo "run-on-job-failure.sh: outcome was $outcome"
fi
