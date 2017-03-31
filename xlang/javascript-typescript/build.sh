#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

type yarn > /dev/null 2>&1 || npm install -g yarn

cd ./buildserver
yarn
./node_modules/.bin/tsc -p .
cd ..

docker build -t ${IMAGE-"xlang-javascript-typescript"} .
