#!/usr/bin/env bash

# This script builds the symbols go binary.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

# Set default empty GOOS
GOOS="${GOOS:-''}"

if [[ "$GOOS" == "linux" ]]; then
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

    # to build the sqlite3 library
    export CC="$muslGcc"
    export CGO_ENABLED=1
fi

echo "--- go build"
go build \
    -buildmode exe \
    -o "$OUTPUT/symbols" \
    github.com/sourcegraph/sourcegraph/cmd/symbols
