#!/usr/bin/env bash

echo "--- deprecated ioutil check"

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit

# Check for ioutil in any .go file, excluding one special case we
# already know about
OUT=$(grep -r ioutil -I --include=\*.go --exclude stdlib.go .)

if [ -z "$OUT" ]; then
  echo "Success: No usages of ioutil found"
  exit 0
else
  echo "ERROR: ioutil check failed"
  echo "$OUT"
  echo "The ioutil package has been deprecated, see: https://golang.org/doc/go1.16#ioutil"
  exit 1
fi
