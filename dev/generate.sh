#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

./cmd/server/generate.sh

go list ./... | grep -v /vendor/ | xargs go generate -v
GO111MODULE=on go mod tidy
