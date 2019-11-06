#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -euxo pipefail

if [ -z "$CTAGS_D_OUTPUT_PATH" ]; then
    echo "buildCtags expects CTAGS_D_OUTPUT_PATH to be set."
    exit 1
fi

cp -R .ctags.d "$CTAGS_D_OUTPUT_PATH"

DOWNLOAD_DIR=`mktemp -d -t sgdockerbuild_XXXXXXX`
cleanup() {
    rm -rf "$DOWNLOAD_DIR"
}
trap cleanup EXIT

curl https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION | tar xz -C $DOWNLOAD_DIR

cd $DOWNLOAD_DIR/ctags-$CTAGS_VERSION
./autogen.sh
LDFLAGS=-static ./configure --program-prefix=universal- --prefix=$OUTPUT_DIR --enable-json --enable-seccomp
make -j8
make install
