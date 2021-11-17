#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../"
set -ex

# We create a file to indicate that this program has already been called
# and there is no need to add a title to the annotation.
FILE=.annotate
LOCKFILE="$FILE.lock"

exec 100>"$LOCKFILE" || exit 1
flock 100 || exit 1

if [ ! -f "$FILE" ]; then
  touch $FILE
  buildkite-agent annotate --context "$BUILDKITE_JOB_ID" --append "$BUILDKITE_LABEL\n"
fi

buildkite-agent annotate --context "$BUILDKITE_JOB_ID" --append "Annotation! 🚀"
