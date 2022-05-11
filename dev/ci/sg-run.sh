#!/usr/bin/env bash

set -e

SG=./bin/sg

if [ -f "$SG" ]; then
  echo "~~~ Found $SG"
  $SG version
else
  echo "~~~ Building $SG"
  go build -o $SG -ldflags "-X main.BuildCommit=$BUILDKITE_COMMIT" -mod=mod .
fi

$SG $@
