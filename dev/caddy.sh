#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

mkdir -p .bin

version="2.4.5"
case "$(go env GOOS)" in
  linux)
    os="linux"
    ;;
  darwin)
    os="mac"
    ;;
esac
name="caddy_${version}_${os}_$(go env GOARCH)"
target="$PWD/.bin/${name}"
url="https://github.com/caddyserver/caddy/releases/download/v${version}/${name}.tar.gz"

if [ ! -f "${target}" ]; then
  echo "downloading ${url}" 1>&2
  curl -sS -L -f "${url}" | tar -xz --to-stdout "caddy" >"${target}.tmp"
  mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd >/dev/null

exec "${target}" "$@"
