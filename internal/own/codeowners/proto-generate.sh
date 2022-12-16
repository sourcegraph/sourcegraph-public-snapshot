#!/usr/bin/env bash

set -euo pipefail

CODEOWNERS="$(readlink -f $(dirname "${BASH_SOURCE[0]}"))"
REPO_ROOT="$(readlink -f "${CODEOWNERS}/../../..")" # cd to repo root dir
cd "${CODEOWNERS}"

GOBIN="${REPO_ROOT}/.bin" go install github.com/bufbuild/buf/cmd/buf
GOBIN="${REPO_ROOT}/.bin" go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
GOBIN="${REPO_ROOT}/.bin" go install golang.org/x/tools/cmd/goimports
GOBIN="${REPO_ROOT}/.bin" go install google.golang.org/protobuf/cmd/protoc-gen-go

GOBIN="${REPO_ROOT}/.bin" "${REPO_ROOT}/.bin/buf" generate "proto/codeowners.proto" --output "${REPO_ROOT}"
"${REPO_ROOT}/.bin/goimports" -w "${CODEOWNERS}/proto/codeowners.pb.go"
