#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null

mkdir -p .bin

version="2.0.0-rc.1"
case "$(go env GOOS)" in
    linux)
        os="Linux"
        ;;
    darwin)
        os="macOS"
        ;;
esac
name="caddy_${version}_${os}_x86_64"
target="$PWD/.bin/${name}"
url="https://github.com/caddyserver/caddy/releases/download/v${version}/${name}.tar.gz"

if [ ! -f "${target}" ]; then
    echo "downloading ${url}" 1>&2
    curl -sS -L -f "${url}" | tar -xz --to-stdout "caddy" > "${target}.tmp"
    mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd > /dev/null

if [ ${SOURCEGRAPH_HTTPS_PORT:-"3443"} -lt 1000 ] && ! [ $(id -u) = 0 ] && hash authbind; then
    # Support using authbind to bind to port 443 as non-root
    exec authbind "${target}" "$@"
else
    exec "${target}" "$@"
fi
