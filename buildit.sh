#!/bin/bash

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

go build ./enterprise/cmd/migrator
