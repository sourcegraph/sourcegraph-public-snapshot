#!/usr/bin/env bash

set -e

echo "--- check codeowners"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

OWNERS_OUT="$(./dev/owners.sh)"

if [ -n "$OWNERS_OUT" ]; then
  echo "$OWNERS_OUT"
  echo "FAILED check: ./dev/owners.sh returned non-empty, indicating there are files without an owner." 1>&2
  exit 1
fi
