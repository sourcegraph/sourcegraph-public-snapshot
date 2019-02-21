#!/bin/bash

set -e

SYMBOLS_IMAGE="${SYMBOLS_IMAGE:=dev-symbols}"
CTAGS_IMAGE="${CTAGS_IMAGE:=ctags}"
BUILD_TYPE="${BUILD_TYPE:=dev}"

repositoryRoot="$PWD"
symbolsExecutablePath="$repositoryRoot/.bin/symbols"
case "$OSTYPE" in
    darwin*)
        libsqlite3PcrePath="$repositoryRoot/libsqlite3-pcre.dylib"
        ;;
    linux*)
        libsqlite3PcrePath="$repositoryRoot/libsqlite3-pcre.so"
        ;;
    *)
        echo "Unknown platform $OSTYPE"
        exit 1
        ;;
esac

# Builds the PCRE extension to sqlite3.
function buildLibsqlite3Pcre() {
    if ! command -v pkg-config >/dev/null 2>&1 || ! command -v pkg-config --cflags sqlite3 libpcre >/dev/null 2>&1; then
        echo "Missing sqlite dependencies."
        case "$OSTYPE" in
            darwin*)
                echo "Install them by running `brew install pkg-config sqlite pcre FiloSottile/musl-cross/musl-cross`"
                ;;
            linux*)
                echo "Install them by running `apt-get install libpcre3-dev libsqlite3-dev pkg-config musl-tools`"
                ;;
            *)
                echo "See the local development documentation: https://github.com/sourcegraph/sourcegraph/blob/master/doc/dev/local_development.md#step-2-install-dependencies"
                ;;
        esac
        exit 1
    fi


    if [ -f "$libsqlite3PcrePath" ]; then
        return
    fi

    sqlite3PcreRepositoryDirectory="$(mktemp -d)"
    trap "rm -rf $sqlite3PcreRepositoryDirectory" EXIT

    echo "Building $libsqlite3PcrePath..."
    curl -fsSL https://codeload.github.com/ralight/sqlite3-pcre/tar.gz/c98da412b431edb4db22d3245c99e6c198d49f7a | tar -C "$sqlite3PcreRepositoryDirectory" -xzvf - --strip 1
    pushd "$sqlite3PcreRepositoryDirectory"
    case "$OSTYPE" in
        darwin*)
            # pkg-config spits out multiple arguments and must not be quoted.
            gcc -fno-common -dynamiclib pcre.c -o "$libsqlite3PcrePath" $(pkg-config --cflags sqlite3 libpcre) $(pkg-config --libs libpcre) -fPIC
            ;;
        linux*)
            # pkg-config spits out multiple arguments and must not be quoted.
            gcc -shared -o "$libsqlite3PcrePath" $(pkg-config --cflags sqlite3 libpcre) -fPIC -W -Werror pcre.c $(pkg-config --libs libpcre) -Wl,-z,defs
            ;;
        *)
            echo "Unknown platform $OSTYPE"
            exit 1
            ;;
    esac
    popd
    echo "Building $libsqlite3PcrePath... done"
}

# Builds the symbols executable.
function buildExecutable() {
    symbolsPackage="github.com/sourcegraph/sourcegraph/cmd/symbols"

    case "$BUILD_TYPE" in
        dev)
            gcFlags="all=-N -l"
            tags="dev delve"
            ;;
        dist)
            gcFlags=""
            tags="dist"
            ;;
    esac

    echo "Building the $symbolsExecutablePath executable..."
    go build -buildmode exe -gcflags "$gcFlags" -tags "$tags" -o "$symbolsExecutablePath" "$symbolsPackage"
    echo "Building the $symbolsExecutablePath executable... done"
}

# Builds and runs the symbols executable.
function execute() {
    buildLibsqlite3Pcre
    buildExecutable
    buildCtagsDockerImage
    export LIBSQLITE3_PCRE="$libsqlite3PcrePath"
    export CTAGS_COMMAND="${CTAGS_COMMAND:=cmd/symbols/universal-ctags-dev}"
    export CTAGS_PROCESSES="${CTAGS_PROCESSES:=1}"
    "$symbolsExecutablePath"
}

# Builds the symbols Docker image.
function buildSymbolsDockerImage() {
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

    symbolsDockerBuildContext="$(mktemp -d)"
    trap "rm -rf $symbolsDockerBuildContext" EXIT

    export CC="$muslGcc"
    export GO111MODULE=on
    export GOARCH=amd64
    export GOOS=linux
    export CGO_ENABLED=1 # to build the sqlite3 library
    symbolsExecutablePath="$symbolsDockerBuildContext/symbols"
    buildExecutable

    buildCtagsDockerImage

    cp -R cmd/symbols/.ctags.d "$symbolsDockerBuildContext"

    echo "Building the $SYMBOLS_IMAGE Docker image..."
    docker build --quiet -f cmd/symbols/Dockerfile -t "$SYMBOLS_IMAGE" "$symbolsDockerBuildContext"
    echo "Building the $SYMBOLS_IMAGE Docker image... done"
}

# Builds the ctags docker image, used by universal-ctags-dev and the symbols Docker image.
function buildCtagsDockerImage() {
    ctagsDockerBuildContext="$(mktemp -d)"
    trap "rm -rf $ctagsDockerBuildContext" EXIT

    cp -R cmd/symbols/.ctags.d "$ctagsDockerBuildContext"

    echo "Building the $CTAGS_IMAGE Docker image..."
    docker build --quiet -f cmd/symbols/internal/pkg/ctags/Dockerfile -t "$CTAGS_IMAGE" "$ctagsDockerBuildContext"
    echo "Building the $CTAGS_IMAGE Docker image... done"
}

command="$1"
if type -t "$command" >/dev/null 2>&1; then
    "$command"
else
    echo "Unknown command: $command (must be: execute, buildLibsqlite3Pcre, or buildSymbolsDockerImage)"
fi
