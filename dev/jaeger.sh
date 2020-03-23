#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null

mkdir -p .bin

version=1.17.1
url="https://github.com/jaegertracing/jaeger/releases/download/v${version}/jaeger-${version}-$(go env GOOS)-$(go env GOARCH).tar.gz"
target="$PWD/.bin/jaeger-all-in-one"

if [ ! -f "${target}" ]; then
    rm -f jaeger.tar.gz
    curl -s -L "${url}" -o "jaeger.tar.gz"
    tar -C "$PWD/.bin/" --strip-components=1 -xzvf jaeger.tar.gz jaeger-1.17.1-linux-amd64/jaeger-all-in-one
    rm -f jaeger.tar.gz
fi

popd > /dev/null

exec "${target}" "$@"
