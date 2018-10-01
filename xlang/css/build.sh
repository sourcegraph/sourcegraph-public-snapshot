#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

if [ ! -d "css-langserver" ]; then
    git clone https://github.com/sourcegraph/css-langserver css-langserver && cd css-langserver/langserver
else
    cd css-langserver && git pull && cd langserver
fi
yarn
./node_modules/.bin/tsc -p .
cd ../..

docker build -t ${IMAGE-"xlang-css"} .
