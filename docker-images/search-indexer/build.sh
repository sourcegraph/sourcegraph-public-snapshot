#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

docker build --squash  -t "${IMAGE:-sourcegraph/zoekt-indexserver}" .
