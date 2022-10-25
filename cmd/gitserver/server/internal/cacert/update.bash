#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"

set -eux -o pipefail

cp "$(go env GOROOT)/src/crypto/x509/root_linux.go" ./
cp "$(go env GOROOT)/src/crypto/x509/root_unix.go" ./
chmod +w root_linux.go
chmod +w root_unix.go

patch -p6 < package-rename.patch
