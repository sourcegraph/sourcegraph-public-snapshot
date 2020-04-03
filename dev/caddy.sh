#!/usr/bin/env bash

set -euf -o pipefail

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null

mkdir -p .bin

beta=20
version="v2.0.0-beta.${beta}"
case "$(go env GOOS)" in
    linux)
        suffix="linux_$(go env GOARCH)"
        ;;
    darwin)
        suffix="macos"
        ;;
esac
suffix="beta${beta}_${suffix}"
target="$PWD/.bin/caddy2_${suffix}"
url="https://github.com/caddyserver/caddy/releases/download/${version}/caddy2_${suffix}"

if [ ! -f "${target}" ]; then
    echo "downloading ${url}" 1>&2
    curl -sS -L -f "${url}" -o "${target}.tmp"
    mv "${target}.tmp" "${target}"
fi

chmod +x "${target}"

popd > /dev/null

if [ ${SOURCEGRAPH_HTTPS_PORT:-"3443"} -lt 1000 ] && ! [ $(id -u) = 0 ] && hash authbind; then
    # Support using authbind to bind to port 443 as non-root
    exec authbind "${target}" "$@"
else
    exec "${$target}" "$@"
fi
