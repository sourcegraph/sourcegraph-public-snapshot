#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/vendor/.bin"
export PATH=$GOBIN:$PATH

go install sourcegraph.com/sourcegraph/sourcegraph/vendor/honnef.co/go/staticcheck/cmd/staticcheck

echo go list...
PKGS=$(go list ./... | grep -v /vendor/)

echo go install...
go install -buildmode=archive ${PKGS}

echo go vet...
go vet ${PKGS}

echo staticcheck...
staticcheck ${PKGS}
