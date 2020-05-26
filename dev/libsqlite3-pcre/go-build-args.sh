#!/usr/bin/env bash

# This script exports variables required to link to the libsqlite3-pcre library.
# This should be sourced prior to the go build command for the binaries that
# require sqlite as a dependency.
#
# Usage: `. ./dev/libsqlite3-pcre/go-build-args.sh`

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

# Set default empty GOOS
GOOS="${GOOS:-''}"

if [[ "$GOOS" != "linux" ]]; then
  exit 0
fi

case "$OSTYPE" in
  darwin*)
    muslGcc="x86_64-linux-musl-gcc"
    if ! command -v "$muslGcc" >/dev/null 2>&1; then
      echo "Couldn't find musl C compiler $muslGcc. Run 'brew install FiloSottile/musl-cross/musl-cross'."
      exit 1
    fi
    ;;

  linux*)
    muslGcc="musl-gcc"
    if ! command -v "$muslGcc" >/dev/null 2>&1; then
      echo "Couldn't find musl C compiler $muslGcc. Install the musl-tools package (e.g. on Ubuntu, run 'apt-get install musl-tools')."
      exit 1
    fi
    ;;

  *)
    echo "Unknown platform $OSTYPE"
    exit 1
    ;;
esac

export CC="$muslGcc"
export CGO_ENABLED=1
