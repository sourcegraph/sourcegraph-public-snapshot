#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

echo "--- yarn in root"
# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --network-timeout 60000

echo "--- cargo install rust-protobuf"
which ./.bin/bin/protoc-gen-rust || cargo install --root .bin protobuf-codegen --version 2.27.1

echo "--- buf"

GOBIN="$PWD/.bin" go install github.com/bufbuild/buf/cmd/buf
GOBIN="$PWD/.bin" go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
GOBIN="$PWD/.bin" go install golang.org/x/tools/cmd/goimports
GOBIN="$PWD/.bin" go install google.golang.org/protobuf/cmd/protoc-gen-go

GOBIN="$PWD/.bin" ./.bin/buf generate
./.bin/goimports -w ./lib/codeintel/lsiftyped/lsif.pb.go

echo "======================"
echo "NOTE:"
echo "If updating SyntaxKind, make sure to update: client/web/src/lsif/spec.ts"
echo "======================"
