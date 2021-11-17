#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

trap "echo ^^^ +++" ERR

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

OUT=$(shfmt -d .)
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
  echo "$OUT" | ./dev/ci/annotate.sh --section "shfmt"
fi
