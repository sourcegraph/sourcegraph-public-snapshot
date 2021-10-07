#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

docker build -t "${IMAGE:-sourcegraph/redis-cache}" .
