#!/usr/bin/env bash

echo "This is for fast testing purposes only. Please test the docker image before submitting changes."
echo
echo "Note: Ensure you have run local-build.sh. This file only updates server, but no dependencies."
echo

cd $(dirname "${BASH_SOURCE[0]}")/../..
export GOBIN=$PWD/cmd/server/.bin
export PATH=$GOBIN:$PATH

export CONFIG_DIR=${CONFIG_DIR:-/tmp/server/etc}
export DATA_DIR=${DATA_DIR:-/tmp/server/data}
echo "CONFIG_DIR=$CONFIG_DIR"
echo "DATA_DIR=$DATA_DIR"
set -ex

type ulimit > /dev/null && ulimit -n 10000 || true

go install -tags dist \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/server

server
