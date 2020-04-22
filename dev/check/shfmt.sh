#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

shfmt -d .
