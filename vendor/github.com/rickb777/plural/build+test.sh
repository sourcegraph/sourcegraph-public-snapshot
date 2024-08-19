#!/bin/bash -e
cd $(dirname $0)
PATH=$HOME/go/bin:$GOPATH/bin:$PATH

if ! type -p goveralls; then
  echo go get github.com/mattn/goveralls
  go get github.com/mattn/goveralls
fi

echo date...
go test -v -covermode=count -coverprofile=date.out .
go tool cover -func=date.out
[ -z "$COVERALLS_TOKEN" ] || goveralls -coverprofile=date.out -service=travis-ci -repotoken $COVERALLS_TOKEN
