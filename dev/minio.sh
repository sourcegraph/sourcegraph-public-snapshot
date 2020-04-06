#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null

mkdir -p .bin

# todo: find a stable URL to download from
version=1.0.0
suffix="$(go env GOOS)-$(go env GOARCH)"
target="$PWD/.bin/minio-${suffix}"
url="https://dl.min.io/server/minio/release/${suffix}/minio"

if [ ! -f "${target}" ]; then
    echo "downloading ${url}" 1>&2
    curl -sS -L -f "${url}" > "${target}.tmp"
    mv "${target}.tmp" "${target}"
    chmod +x "${target}"
fi

chmod +x "${target}"

popd > /dev/null

datadir="~/.sourcegraph/blobs"

export MINIO_ACCESS_KEY=sourcegraph
export MINIO_SECRET_KEY=sourcegraph

exec "${target}" server $datadir "$@"
