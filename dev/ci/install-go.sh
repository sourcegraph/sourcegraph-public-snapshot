#!/bin/bash

set -ex

TARBALL=go1.7.linux-amd64.tar.gz

mkdir -p ~/cache
cd ~/cache
[ -f $TARBALL ] || curl -O https://storage.googleapis.com/golang/$TARBALL

cd /usr/local
sudo rm -rf go
sudo tar -xzf $HOME/cache/$TARBALL
sudo chmod -R a+rwx /usr/local/go
