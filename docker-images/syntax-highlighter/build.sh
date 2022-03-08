#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Make sure we have all tree-sitter modules cloned.
git submodule update --init --recursive

docker build -t "${IMAGE:-sourcegraph/syntax-highlighter}" .
