#!/usr/bin/env bash

echo "This is for fast testing purposes only. Please test the docker image before submitting changes."
echo
echo "Note: Ensure you have run local-build.sh. This file only updates server, but no dependencies."
echo

cd $(dirname "${BASH_SOURCE[0]}")/../..
export GOBIN=$PWD/cmd/server/.bin
export PATH=$GOBIN:$PATH
set -ex

go install -tags dist \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/server

server
