#!/usr/bin/env bash

set -e

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# setup additional gitignore source
awk -F " +@" '{ print $1 }' <.github/CODEOWNERS >/tmp/ignore

set +e
find . -path ./.git -prune -o -print | git -c core.excludesfile=/tmp/ignore check-ignore --verbose --no-index -n --stdin | grep '^::\s' | grep -v '^::\s\.$'
set -e
