#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

set -e
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set +e
OUT=$(shfmt -d .)
EXIT_CODE=$?
set -e
echo -e "$OUT"

if [ $EXIT_CODE -ne 0 ]; then
  echo -e "$OUT" | ./enterprise/dev/ci/scripts/annotate.sh -s "shfmt"
  echo "^^^ +++"
fi

exit $EXIT_CODE
