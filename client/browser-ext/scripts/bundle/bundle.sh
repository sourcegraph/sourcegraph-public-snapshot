#!/bin/bash

set -e

cd /browser-ext

rm -rf node_modules
rm -rf build

yarn install
yarn run build

rm -f firefox-bundle.xpi
rm -f chrome-bundle.zip

zip -r firefox-bundle.xpi build/*
zip -r chrome-bundle.zip build/*