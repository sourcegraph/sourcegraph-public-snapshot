#!/usr/bin/env bash

echo "--- shellcheck"

trap "echo ^^^ +++" ERR

set -e

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

SHELL_SCRIPTS=()

# ignore dev/sg/internal/usershell/autocomplete which just houses scripts copied from elsewhere
IGNORE_AUTOCOMPLETE="dev/sg/internal/usershell/autocomplete"
# ignore client/jetbrains since the shell scripts are created by gradle and not maintained by us
IGNORE_JETBRAINS="client/jetbrains"
# ignore scip-syntax as gradle scripts are generated
IGNORE_SCIP_SYNTAX="docker-images/syntax-highlighter/crates/scip-syntax"

GREP_IGNORE_FILES="$IGNORE_AUTOCOMPLETE\|$IGNORE_JETBRAINS\|$IGNORE_SCIP_SYNTAX"

while IFS='' read -r line; do SHELL_SCRIPTS+=("$line"); done < <(comm -12 <(git ls-files | sort) <(shfmt -f . | grep -v $GREP_IGNORE_FILES | sort))

set +e
OUT=$(shellcheck --external-sources --source-path="SCRIPTDIR" --color=always "${SHELL_SCRIPTS[@]}")
EXIT_CODE=$?
set -e
echo -e "$OUT"

if [ $EXIT_CODE -ne 0 ]; then
  mkdir -p ./annotations
  echo -e "$OUT" >./annotations/shellcheck
  echo "^^^ +++"
fi

exit $EXIT_CODE
