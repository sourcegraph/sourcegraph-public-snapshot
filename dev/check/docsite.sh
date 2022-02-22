#!/usr/bin/env bash

set -e

echo "--- docsite check (lint Markdown files in doc/)"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

# Check broken links, etc., in Markdown files in doc/.

set +e
OUT=$(./dev/docsite.sh check)
EXIT_CODE=$?
set -e

echo -e "$OUT"

if [ $EXIT_CODE -ne 0 ] && [[ "$CI" == "true" ]]; then
  echo -e "$OUT" >./annotations/docsite
  echo "^^^ +++"
fi

exit "$EXIT_CODE"
