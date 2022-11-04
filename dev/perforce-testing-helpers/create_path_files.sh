#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -euxo pipefail

cd base

find . -type f -name "path.txt" -exec rm -- {} \;

find . -type d -exec sh -c 'dir="$1"; cd "$dir"; echo "$dir/" >path.txt' bash {} \;
