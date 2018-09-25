#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

rm -f ./cmd/server/dockerfile.go
cp ../sourcegraph/cmd/server/dockerfile.go ./cmd/server/dockerfile.go

SERVER_PKG=github.com/sourcegraph/enterprise/cmd/server ../sourcegraph/cmd/server/build.sh \
    github.com/sourcegraph/enterprise/cmd/frontend
