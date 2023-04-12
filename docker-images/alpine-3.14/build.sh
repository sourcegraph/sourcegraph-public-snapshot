#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

docker build --platform linux/amd64 -t "${IMAGE:-sourcegraph/alpine-3.14}" .
