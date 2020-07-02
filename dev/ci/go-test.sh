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

goacc=""
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --goacc)
      goacc="$2"
      shift
      ;;
    *)
      echo "Unknown parameter passed: $1"
      exit 1
      ;;
  esac
  shift
done

if [ "$goacc" == "true" ]; then
  echo "--- go test with accurate code coverage"
  go get github.com/ory/go-acc
  "$(go env GOPATH)/bin/go-acc" ./...
else
  echo "--- go test"
  go test -timeout 4m -coverprofile=coverage.txt -covermode=atomic -race ./...
fi
