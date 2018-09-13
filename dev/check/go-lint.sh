#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/vendor/.bin"
export PATH=$GOBIN:$PATH

go install github.com/sourcegraph/sourcegraph/vendor/honnef.co/go/tools/cmd/staticcheck

echo go list...
PKGS=$(go list ./... | grep -v /vendor/)

echo go install...
go install -buildmode=archive ${PKGS}

echo go vet...
go vet ${PKGS}

echo staticcheck...
staticcheck -ignore 'github.com/sourcegraph/sourcegraph/cmd/gitserver/server/server_test.go:SA1019 github.com/sourcegraph/sourcegraph/pkg/httptestutil/client.go:SA1019' ${PKGS}
