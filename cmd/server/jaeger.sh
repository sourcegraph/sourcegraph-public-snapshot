#!/usr/bin/env bash

set -euf -o pipefail

[ -n "$OUTPUT" ]

version=1.24.0
suffix="${version}-$(go env GOOS)-$(go env GOARCH)"
target="${OUTPUT}/usr/local/bin/jaeger"
url="https://github.com/jaegertracing/jaeger/releases/download/v${version}/jaeger-${suffix}.tar.gz"

mkdir -p "$(dirname "$target")"
curl -sS -L -f "${url}" | tar -xz --to-stdout "jaeger-${suffix}/jaeger-all-in-one" >"${target}.tmp"
mv "${target}.tmp" "${target}"

chmod +x "${target}"
