#!/usr/bin/env bash

set -euxo pipefail

set +u
if [ -z "$CC" ]; then
    set -u

    case "$OSTYPE" in
        darwin*)
            muslGcc="x86_64-linux-musl-gcc"
            if ! command -v "$muslGcc" >/dev/null 2>&1; then
                echo "Couldn't find musl C compiler $muslGcc. Run `brew install FiloSottile/musl-cross/musl-cross`."
                exit 1
            fi
            ;;
        linux*)
            muslGcc="musl-gcc"
            if ! command -v "$muslGcc" >/dev/null 2>&1; then
                echo "Couldn't find musl C compiler $muslGcc. Install the musl-tools package (e.g. on Ubuntu, run `apt-get install musl-tools`)."
                exit 1
            fi
            ;;
        *)
            echo "Unknown platform $OSTYPE"
            exit 1
            ;;
    esac

    export CC="$muslGcc"
fi

set -u

curl -fsSL https://codeload.github.com/ralight/sqlite3-pcre/tar.gz/c98da412b431edb4db22d3245c99e6c198d49f7a | tar -C $OUTPUT_DIR -xzvf - --strip 1
cd $OUTPUT_DIR
make CC="$CC"
