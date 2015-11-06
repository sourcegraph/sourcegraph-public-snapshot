#!/bin/bash

set -e

TMPDIR=$(mktemp -dq)

cd "$TMPDIR"
wget -q https://github.com/google/protobuf/archive/v3.0.0-alpha-3.tar.gz
tar zxf v3.0.0-alpha-3.tar.gz
cd protobuf-3.0.0-alpha-3
./autogen.sh
./configure
make -j2 install DESTDIR="$1"
cd
rm -rf "$TMPDIR/protobuf-*"
