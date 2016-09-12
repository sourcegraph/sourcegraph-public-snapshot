#!/bin/bash

set -ex

VERSION="v0.1.0"
mkdir -p ~/cache/checkup
cd ~/cache/checkup
[ -f "checkup_linux_amd64-${VERSION}.zip" ] || curl -Lk https://github.com/sourcegraph/checkup/releases/download/$VERSION/checkup_linux_amd64.zip > "checkup_linux_amd64-${VERSION}.zip"

unzip -o "checkup_linux_amd64-${VERSION}.zip"
sudo mv checkup /usr/local/bin/checkup
