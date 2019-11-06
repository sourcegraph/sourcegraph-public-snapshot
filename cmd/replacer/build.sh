#!/usr/bin/env bash

# This function installs the Comby dependency for replacer. It is called when
# this script is called with 'installComby' in the first argument (e.g., by CI).  
# If this script is called without arguments, it simply builds replacer.
function installComby() {
    RELEASE_VERSION="0.11.0"
    RELEASE_TAG="0.11.0"
    RELEASE_URL="https://github.com/comby-tools/comby/releases"

    INSTALL_DIR=/usr/local/bin

    function ctrl_c() {
        rm -f $TMP/$RELEASE_BIN &> /dev/null
        printf "[-] Installation cancelled. Please see https://github.com/comby-tools/comby/releases if you prefer to install manually.\n"
        exit 1
    }

    trap ctrl_c INT

    EXISTS=$(command -v comby || echo)

    if [ -n "$EXISTS" ]; then
        INSTALL_DIR=$(dirname $EXISTS)
    fi

    if [ ! -d "$INSTALL_DIR" ]; then
        printf "$INSTALL_DIR does not exist. Please download the binary from ${RELEASE_URL} and install it manually.\n"
        exit 1
    fi

    TMP=${TMPDIR:-/tmp}

    ARCH=$(uname -m || echo)
    case "$ARCH" in
        x86_64|amd64) ARCH="x86_64";;
        *) ARCH="OTHER"
    esac

    OS=$(uname -s || echo)
    if [ "$OS" = "Darwin" ]; then
        OS=macos
    fi

    RELEASE_BIN="comby-${RELEASE_TAG}-${ARCH}-${OS}"
    RELEASE_URL="https://github.com/comby-tools/comby/releases/download/${RELEASE_TAG}/${RELEASE_BIN}"

    if [ ! -e "$TMP/$RELEASE_BIN" ]; then
        printf "[+] Downloading comby $RELEASE_VERSION\n"

        SUCCESS=$(curl -s -L -o "$TMP/$RELEASE_BIN" "$RELEASE_URL" --write-out "%{http_code}")

        if [ "$SUCCESS" == "404" ]; then
            printf "[-] No binary release available for your system.\n"
            rm -f $TMP/$RELEASE_BIN
            exit 1
    fi
	printf "[+] Download complete.\n"
    fi

    chmod 755 "$TMP/$RELEASE_BIN"
    echo "[+] Installing comby to $INSTALL_DIR"
    if [ ! $OS == "macos" ]; then
        sudo cp "$TMP/$RELEASE_BIN" "$INSTALL_DIR/comby"
    else
        cp "$TMP/$RELEASE_BIN" "$INSTALL_DIR/comby"
    fi

    SUCCESS_IN_PATH=$(command -v comby || echo notinpath)

    if [ $SUCCESS_IN_PATH == "notinpath" ]; then
        printf "[*] Comby is not in your PATH. You should add $INSTALL_DIR to your PATH.\n"
        rm -f $TMP/$RELEASE_BIN
        exit 1
    fi

    CHECK=$(printf 'printf("hello world!\\\n");' | $INSTALL_DIR/comby 'printf("hello :[1]!\\n");' 'printf("hello comby!\\n");' -stdin || echo broken)
    if [ "$CHECK"  == "broken" ]; then
        printf "[-] comby did not install correctly.\n"
        printf "[-] My guess is that you need to install the pcre library on your system. Try:\n"
        if [ $OS == "macos" ]; then
            printf "brew install pcre && bash <(curl -sL get.comby.dev)$\n"
        else
            printf "[*]sudo apt-get install libpcre3-dev && bash <(curl -sL get.comby.dev)\n"
        fi
        rm -f $TMP/$RELEASE_BIN
        exit 1
    fi

    rm -f $TMP/$RELEASE_BIN
}

command="$1"
if [ "$command" = "installComby" ]; then
    installComby
    exit 0
fi

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

OUTPUT=`mktemp -d -t sgdockerbuild_XXXXXXX`
cleanup() {
    rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

for pkg in github.com/sourcegraph/sourcegraph/cmd/replacer; do
    go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION" -buildmode exe -tags dist -o $OUTPUT/$(basename $pkg) $pkg
done

docker build -f cmd/replacer/Dockerfile -t $IMAGE $OUTPUT \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
