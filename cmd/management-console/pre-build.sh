#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -ex

pushd web/
npm install
npm run build
popd
go generate ./assets
