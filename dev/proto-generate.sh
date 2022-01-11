#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

GOBIN="$PWD/.bin" go install golang.org/x/tools/cmd/goimports
GOBIN="$PWD/.bin" go install github.com/bufbuild/buf/cmd/buf
GOBIN="$PWD/.bin" go install google.golang.org/protobuf/cmd/protoc-gen-go

GOBIN="$PWD/.bin" ./.bin/buf generate
./.bin/goimports -w ./lib/codeintel/lsif_typed/lsif.pb.go
