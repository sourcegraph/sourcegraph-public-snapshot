#!/usr/bin/env bash

# build src-cli on macOS

cd "${1:-${HOME}/Downloads}" || exit 1

[ -d src-cli ] || git clone https://github.com/sourcegraph/src-cli.git
cd src-cli || exit 1
git reset --hard
git clean -dfxq
git pull
for arch in arm64 amd64
do
    GOARCH=${arch} go build -o src-${arch} ./cmd/src
done
lipo src-{arm64,amd64} -create -output src-universal
echo "${PWD}/src-universal"
exit 0
