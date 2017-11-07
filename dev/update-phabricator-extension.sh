#!/bin/bash

set -e

cd "$(dirname "${BASH_SOURCE[0]}")/../client/phabricator" # cd to repo root dir

rm -rf .extension
git clone git@github.com:sourcegraph/browser-extension.git .extension
cd .extension

npm install
npm run build

cp dist/js/phabricator.bundle.js ../scripts
cp dist/css/style.bundle.css ../scripts

cd ..
rm -rf .extension
cd ../..

mkdir -p ui/assets/extension
mkdir -p ui/assets/extension/scripts
mkdir -p ui/assets/extension/css
cp ./client/phabricator/scripts/phabricator.bundle.js ui/assets/extension/scripts
cp ./client/phabricator/scripts/style.bundle.css ui/assets/extension/css
