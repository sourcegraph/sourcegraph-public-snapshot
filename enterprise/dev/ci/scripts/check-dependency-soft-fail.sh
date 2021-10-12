#!/usr/bin/env bash

# This script runs the given build command if the given the dependency build step's outcome
# was not a soft_fail.

DEPENDENCY_BUILD_STEP=$1

set -ex -o pipefail

outcome=$(buildkite-agent step get "outcome" --step "$DEPENDENCY_BUILD_STEP")

if [ "$outcome" != "soft_failed" ]; then
  echo "+++ $DEPENDENCY_BUILD_STEP did not exit with soft_fail, exited with $outcome instead"
  exit 0
else
  echo "+++ $DEPENDENCY_BUILD_STEP exited with soft_fail, exiting 222 soft fail"
  exit 222
fi
