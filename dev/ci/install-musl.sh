#!/bin/bash

set -ex

wget http://www.musl-libc.org/releases/musl-1.1.10.tar.gz
tar -xvf musl-1.1.10.tar.gz
cd musl-1.1.10
./configure
make
sudo make install
cd ..
rm -rf musl-1.1.10 musl-1.1.10.tar.gz
