#!/usr/bin/env bash

# Exits with code zero if the need for rebuild of the executors AMIs is detected.
#
# Usage if [ ./ci-should-rebuild.sh.sh ]; then ./your/command; fi

set -eu

# target-determinator looks at all the deps for the given targets and only print them back on STDOOUT
# if it finds that one of inputs (transitive or not) has changed since a given commit.
#
# It effectively computes if yes or no something changed since the given commit.
changes_count=$(target-determinator \
  --verbose \
  --targets \
  "//cmd/executor/vm-image:ami.build union //cmd/executor/docker-mirror:ami.build" \
  HEAD^ \
  | wc -l
)

if [[ "$changes_count" -eq 0 ]]; then
  exit 1
fi
