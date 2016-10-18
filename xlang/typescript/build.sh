#!/bin/bash
set -ex

if [ ! -d "langserver" ]; then
    git clone https://github.com/sourcegraph/langserver-typescript langserver && cd langserver
else
    cd langserver && git pull
fi

npm install
./node_modules/.bin/tsc -p .
cd ..
