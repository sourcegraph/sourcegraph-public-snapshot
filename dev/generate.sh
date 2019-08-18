#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

# Generate these first because they are necessary for the other packages to compile.
go generate ./migrations ./cmd/frontend/graphqlbackend

go list ./... | xargs go generate -x
