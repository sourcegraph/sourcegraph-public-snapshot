#!/bin/bash

set -ex

if [ -d ~/musl ]; then
    sudo ln -sf ~/musl /usr/local/musl
    exit 0
fi

cd ~
wget http://www.musl-libc.org/releases/musl-1.1.10.tar.gz
tar -xvf musl-1.1.10.tar.gz
cd musl-1.1.10
./configure
make
sudo make install
cd ..
rm -rf musl-1.1.10 musl-1.1.10.tar.gz

# Copy musl directory to the home directory so we can cache it
cp -r /usr/local/musl musl
