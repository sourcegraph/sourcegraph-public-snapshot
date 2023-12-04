#!/usr/bin/env bash

set -u

response=$(buildkite-agent artifact download bazel-configure.diff . --step bazel-prechecks 2>&1)
status=$?
if [[ $status -ne 0 && "$response" == *"No artifacts found for downloading"* ]]; then
  echo "--- No bazel-configure.diff artifact found, skipping diff check"
  exit 0
fi

if [ -f bazel-configure.diff ]; then
  git apply bazel-configure.diff
fi
