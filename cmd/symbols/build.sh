#!/usr/bin/env bash

set -exo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

BUILD_TYPE="${BUILD_TYPE:=dev}"

repositoryRoot="$PWD"
SYMBOLS_EXECUTABLE_OUTPUT_PATH="${SYMBOLS_EXECUTABLE_OUTPUT_PATH:=$repositoryRoot/.bin/symbols}"
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

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

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

    if [ "$GOOS" = "linux" ]; then
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
        export CGO_ENABLED=1 # to build the sqlite3 library
    fi

    echo "Building the $SYMBOLS_EXECUTABLE_OUTPUT_PATH executable..."
    go build -buildmode exe -gcflags "$gcFlags" -tags "$tags" -o "$SYMBOLS_EXECUTABLE_OUTPUT_PATH" "$symbolsPackage"
    echo "Building the $SYMBOLS_EXECUTABLE_OUTPUT_PATH executable... done"
}

export -f buildExecutable

# Builds and runs the symbols executable.
function execute() {
    buildLibsqlite3Pcre
    buildExecutable
    env IMAGE=ctags ./cmd/symbols/internal/pkg/ctags/docker.sh
    export LIBSQLITE3_PCRE="$libsqlite3PcrePath"
    export CTAGS_COMMAND="${CTAGS_COMMAND:=cmd/symbols/universal-ctags-dev}"
    export CTAGS_PROCESSES="${CTAGS_PROCESSES:=1}"
    "$SYMBOLS_EXECUTABLE_OUTPUT_PATH"
}

# Builds the Docker images that the symbols Docker image depends on. The caller
# must set:
#
# - SYMBOLS_EXECUTABLE_OUTPUT_PATH
# - CTAGS_D_OUTPUT_PATH
function buildSymbolsDockerImageDependencies() {
    if [ -z "$SYMBOLS_EXECUTABLE_OUTPUT_PATH" ]; then
        echo "buildSymbolsDockerImageDependencies expects SYMBOLS_EXECUTABLE_OUTPUT_PATH to be set."
        exit 1
    fi
    if [ -z "$CTAGS_D_OUTPUT_PATH" ]; then
        echo "buildSymbolsDockerImageDependencies expects CTAGS_D_OUTPUT_PATH to be set."
        exit 1
    fi

    export GO111MODULE=on
    export GOARCH=amd64
    export GOOS=linux

    echo "--- build symbols dependencies"
    parallel_run {} ::: "cmd/symbols/internal/pkg/ctags/build.sh" "cmd/symbols/libsqlite3-pcre/build.sh"

    buildExecutable

    cp -R cmd/symbols/.ctags.d "$CTAGS_D_OUTPUT_PATH"
}

# TODO@ggilmore: This is ugly, but this works for now and
# until the symbols build file is refactored into a makefile
function build_binary_and_universal_ctags (){
    if [ -z "$SYMBOLS_EXECUTABLE_OUTPUT_PATH" ]; then
        echo "build_binary_and_universal_ctags expects SYMBOLS_EXECUTABLE_OUTPUT_PATH to be set."
        exit 1
    fi
    if [ -z "$CTAGS_D_OUTPUT_PATH" ]; then
        echo "build_binary_and_universal_ctags expects CTAGS_D_OUTPUT_PATH to be set."
        exit 1
    fi

    export GO111MODULE=on
    export GOARCH=amd64
    export GOOS=linux

    parallel_run {} ::: buildExecutable "cmd/symbols/internal/pkg/ctags/build.sh"

    cp -R cmd/symbols/.ctags.d "$CTAGS_D_OUTPUT_PATH"
}

# Builds the symbols Docker image.
function buildSymbolsDockerImage() {
    SYMBOLS_EXECUTABLE_OUTPUT_PATH="$OUTPUT_DIR/symbols"
    CTAGS_D_OUTPUT_PATH="$OUTPUT_DIR/.ctags.d"

    buildSymbolsDockerImageDependencies
}

command="$1"
if type -t "$command" >/dev/null 2>&1; then
    "$command"
else
    echo "Unknown command: $command (must be: execute, buildLibsqlite3Pcre, or buildSymbolsDockerImage)"
fi
