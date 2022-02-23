#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

set -e
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set +e

# Ignore scripts in submodules
#
# Following command uses `git ls-files` to make sure we have only files that
# are in our repository, and then `shfmt -f .` to find the shell scripts.
#
# `comm` is used to find the common items between the two
OUT=$(comm -12 <(git ls-files | sort) <(shfmt -f . | sort) | shfmt -d)
EXIT_CODE=$?
set -e
echo -e "$OUT"

if [ $EXIT_CODE -ne 0 ]; then
  echo -e "$OUT" >./annotations/shfmt
  echo "^^^ +++"
fi

exit $EXIT_CODE
