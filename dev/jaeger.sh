#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null

mkdir -p .bin

version=1.17.1
suffix="${version}-$(go env GOOS)-$(go env GOARCH)"
target="$PWD/.bin/jaeger-all-in-one-${suffix}"
url="https://github.com/jaegertracing/jaeger/releases/download/v${version}/jaeger-${suffix}.tar.gz"

if [ ! -f "${target}" ]; then
    echo "downloading ${url}" 1>&2
    curl -sS -L -f "${url}" | tar -x --to-stdout "jaeger-${suffix}/jaeger-all-in-one" > "${target}.tmp"
    mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd > /dev/null

exec "${target}" --log-level "${JAEGER_LOG_LEVEL:-info}" "$@"
