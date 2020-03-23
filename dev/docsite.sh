#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null

mkdir -p .bin

version=v1.4.0
binname="docsite_${version}_$(go env GOOS)_$(go env GOARCH)"
target="$PWD/.bin/${binname}"

if [ ! -f "${target}" ]; then
    curl -s -L "https://github.com/sourcegraph/docsite/releases/download/${version}/${binname}" -o "${target}.tmp"
    mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd > /dev/null

exec "${target}" "$@"
