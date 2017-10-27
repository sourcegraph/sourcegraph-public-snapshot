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

rm -rf .extension
