#!/usr/bin/env bash

set -e

# For symbol tests
echo "--- build libsqlite"
./dev/libsqlite3-pcre/build.sh

# For searcher and replacer tests
echo "--- comby install"
./dev/comby-install-or-upgrade.sh

# Separate out time for go mod from go test
echo "--- go mod download"
go mod download

echo "--- go coverage"
# From https://github.com/ory/go-acc
touch ./coverage.tmp
echo 'mode: atomic' >coverage.txt
# shellcheck disable=SC2016
go list ./... | grep -v /vendor | xargs -n1 -I{} sh -c 'go test -covermode=atomic -coverprofile=coverage.tmp -coverpkg $(go list ./... | grep -v /vendor | tr "\n" ",") {} && tail -n +2 coverage.tmp >> coverage.txt || exit 255' && rm coverage.tmp

echo "--- go test"
go test -timeout 4m -race ./...
