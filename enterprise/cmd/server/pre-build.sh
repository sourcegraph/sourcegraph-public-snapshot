#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

./dev/generate.sh
./cmd/frontend/pre-build.sh
