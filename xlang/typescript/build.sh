#!/bin/bash
set -ex

if [ ! -d "langserver" ]; then
    git clone https://github.com/sourcegraph/langserver langserver && cd langserver
else
    cd langserver && git pull
fi

npm install
tsc -p .
cd ..
