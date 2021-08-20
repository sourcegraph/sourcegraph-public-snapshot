#!/usr/bin/env bash

set -euo pipefail

asdf install
yarn
yarn generate
echo yes | gcloud auth configure-docker

curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
