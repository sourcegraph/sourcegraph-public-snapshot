#!/bin/bash
set -ex

go get -d golang.org/x/tools/cmd/guru/...
GOOS=linux GOARCH=amd64 go build -o guru golang.org/x/tools/cmd/guru
GOOS=linux GOARCH=amd64 go build -o langserver-go .

docker build -t us.gcr.io/sourcegraph-dev/langserver-go .
gcloud docker push us.gcr.io/sourcegraph-dev/langserver-go
