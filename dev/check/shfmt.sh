#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

OUT=$(shfmt -d .)
EXIT_CODE=$?
echo "$OUT"

if [ $EXIT_CODE -ne 0 ]; then
  echo "$OUT" | ./dev/ci/annotate.sh --section "shfmt"
fi

exit $EXIT_CODE
