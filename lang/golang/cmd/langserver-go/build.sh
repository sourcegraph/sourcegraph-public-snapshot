#!/bin/bash
set -ex

go get -u -d golang.org/x/tools/cmd/guru/... github.com/rogpeppe/godef/... sourcegraph.com/sourcegraph/srclib-go/gog/...
GOOS=linux GOARCH=amd64 go build -o guru golang.org/x/tools/cmd/guru
GOOS=linux GOARCH=amd64 go build -o godef github.com/rogpeppe/godef
GOOS=linux GOARCH=amd64 go build -o gog sourcegraph.com/sourcegraph/srclib-go/gog/cmd/gog
GOOS=linux GOARCH=amd64 go build -o langserver-go .

docker build -t us.gcr.io/sourcegraph-dev/langserver-go .
gcloud docker push us.gcr.io/sourcegraph-dev/langserver-go
