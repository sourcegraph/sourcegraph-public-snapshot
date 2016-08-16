#!/bin/bash

set -ex

cd /usr/local
sudo rm -rf go
curl https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz | sudo tar -xz
sudo chmod -R a+rwx /usr/local/go
