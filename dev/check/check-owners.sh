#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

OWNERS_OUT="$(./dev/owners)"
if [ ! -z "$OWNERS_OUT" ]; then
    echo "$OWNERS_OUT"
    echo "FAILED check: ./dev/owners returned non-empty, indicating there are files without an owner." 1>&2
    exit 1
fi
