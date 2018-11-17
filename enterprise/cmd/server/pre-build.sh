#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

cp ../cmd/server/dockerfile.go cmd/server/
./cmd/frontend/pre-build.sh
