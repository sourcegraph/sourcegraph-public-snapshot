#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

mkdir -p .bin

version="1.50.1"
suffix="${version}-$(go env GOOS)-$(go env GOARCH)"
name="golangci-lint-${suffix}"
target="$PWD/.bin/${name}"
url="https://github.com/golangci/golangci-lint/releases/download/v$version/golangci-lint-$suffix.tar.gz"

if [ ! -f "${target}" ]; then
  echo "installing golangci-lint" 1>&2
  curl -sS -L -f "${url}" | tar -xz --to-stdout "${name}/golangci-lint" >"${target}.tmp"
  mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd >/dev/null

exec "${target}" "$@"
