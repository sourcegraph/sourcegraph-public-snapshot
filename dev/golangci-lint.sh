#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

mkdir -p .bin

version="1.45.2"
suffix="${version}-$(go env GOOS)-$(go env GOARCH)"
target="$PWD/.bin/golangci-lint-${suffix}"

if [ ! -f "${target}" ]; then
  # Workaround for https://github.com/golangci/golangci-lint/issues/2374
  echo "installing golangci-lint" 1>&2
  go install "github.com/golangci/golangci-lint/cmd/golangci-lint@v${version}" 1>&2
  mv "$(which golangci-lint)" "${target}"
fi

chmod +x "${target}"

popd >/dev/null

exec "${target}" "$@"
