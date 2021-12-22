#!/usr/bin/env bash

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/../../.." >/dev/null

trap 'rm dev/sg/sg_docker' EXIT

pushd dev/sg
GOOS=linux GOARCH=amd64 go build -o sg_docker .
popd

docker build -t sg:test . --platform linux/amd64 -f dev/sg/testing/Dockerfile
