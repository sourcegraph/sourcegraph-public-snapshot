#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

cd "$(dirname "${BASH_SOURCE[0]}")"/../.. || exit

OUT=$(shfmt -d .)
EXIT_CODE=$?
echo -e "$OUT"

if [ $EXIT_CODE -ne 0 ]; then
  echo -e "$OUT" | ./dev/ci/annotate.sh -s "shfmt"
  echo "^^^ +++"
fi

exit $EXIT_CODE
