#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -euxo pipefail

if [ ! $(which fd) ]; then
    echo "'fd' command not found. Please install 'fd' from https://github.com/sharkdp/fd"
    exit 1
fi

if [ ! $(which sd) ]; then
    echo "'sd' command not found. Please install 'sd' from https://github.com/chmln/sd"
    exit 1
fi

COMMIT="$1"
DIGEST="$2"

TAG="$COMMIT@sha256:$DIGEST"

for file in $(fd 'Dockerfile'); do
    sd "FROM sourcegraph/builder(.*) as (.*)" "FROM sourcegraph/builder:${TAG} as \$2" $file
done
