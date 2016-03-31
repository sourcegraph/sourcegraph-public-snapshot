#!/bin/bash

set -e

pkgs=(git curl libcurl3 libcurl3-gnutls)

# Work from the directory CI will cache
mkdir -p ~/cache/deb
cd ~/cache/deb

# check we have a deb for each package
useCache=true
for pkg in "${pkgs[@]}"; do
    if ! ls | grep "^${pkg}"; then
        useCache=false
    fi
done

set -x

if [ ${useCache} == true ]; then
    sudo dpkg -i *.deb
    exit 0
fi

sudo apt-get update
sudo apt-get install --reinstall "${pkgs[@]}"

cp -v /var/cache/apt/archives/*.deb ~/cache/deb/
