#!/usr/bin/env bash

echo "--- shellcheck"

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

SHELL_SCRIPTS=()

while IFS='' read -r line; do SHELL_SCRIPTS+=("$line"); done < <(shfmt -f .)

OUT=$(shellcheck --external-sources --source-path="SCRIPTDIR" --color=always "${SHELL_SCRIPTS[@]}")
EXIT_CODE=$?
echo -e $OUT

if [ $EXIT_CODE -ne 0 ]; then
  echo -e "$OUT" | ./dev/ci/annotate.sh -s "shellcheck"
  echo "^^^ +++"
fi

exit $EXIT_CODE
