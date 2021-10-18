#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

version=v1.8.2
suffix="${version}_$(go env GOOS)_$(go env GOARCH)"
url="https://github.com/sourcegraph/docsite/releases/download/${version}/docsite_${suffix}"

base="$PWD/.bin"
if [[ "${CI:-"false"}" == "true" ]]; then
  base="/tmp"
fi

target="${base}/docsite_${suffix}"
mkdir -p "$(dirname "${target}")"

if [ ! -f "${target}" ]; then
  echo "downloading ${url}" 1>&2
  curl -sS -L -f "${url}" -o "${target}.tmp"
  mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd >/dev/null

exec "${target}" "$@"
