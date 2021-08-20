#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

mkdir -p .bin

version="1.37.1"
suffix="${version}-$(go env GOOS)-$(go env GOARCH)"
target="$PWD/.bin/golangci-lint-${suffix}"
url="https://github.com/golangci/golangci-lint/releases/download/v${version}/golangci-lint-${suffix}.tar.gz"

if [ ! -f "${target}" ]; then
  echo "downloading ${url}" 1>&2
  curl -sS -L -f "${url}" -o "${target}.tar.gz"
  tar xzf "${target}.tar.gz"
  mv "golangci-lint-${suffix}/golangci-lint" "${target}"
  rm -f "${target}.tar.gz"
  rm -rf "golangci-lint-${suffix}"
fi

chmod +x "${target}"

popd >/dev/null

exec "${target}" "$@"
