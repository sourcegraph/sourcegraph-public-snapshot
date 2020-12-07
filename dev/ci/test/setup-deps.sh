#!/usr/bin/env bash

asdf install
yarn
yarn generate
yes | gcloud auth configure-docker

apt-get install -y gcc python3-dev python3-setuptools
pip3 install --no-cache-dir -U crcmod

curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
