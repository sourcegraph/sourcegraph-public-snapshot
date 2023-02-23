#!/usr/bin/env bash

cd "${1:-${HOME}/Downloads}" || exit 1

[ -d zoekt ] || git clone https://github.com/sourcegraph/zoekt.git
cd zoekt || exit 1
git reset --hard
git clean -dfxq
git pull
for service in webserver indexserver
do
    for arch in arm64 amd64
    do
        GOARCH=${arch} go build -o zoekt-${service}-${arch} github.com/sourcegraph/zoekt/cmd/zoekt-${service}
    done
    lipo zoekt-${service}-{arm64,amd64} -create -output zoekt-${service}-universal
done
