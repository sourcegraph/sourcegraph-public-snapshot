#!/bin/bash

set -e
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

for dir in web shared packages/sourcegraph-extension-api client/browser; do
    (set -x; cd "$dir" && "$@")
done
