#!/bin/sh

set -eux

CTAGS_VERSION=03f933a96d3ef87adbf9d167462d45ce69577edb

apk --no-cache add --virtual build-deps curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils
curl https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION | tar xz -C /tmp

cd /tmp/ctags-$CTAGS_VERSION
./autogen.sh
LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp
make -j8
make install
cd -

rm -rf /tmp/ctags-$CTAGS_VERSION
