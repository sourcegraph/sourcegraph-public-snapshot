#!/bin/bash

set -e # quit script if anything fails
cd /browser-ext
rm -rf node_modules
rm -rf build/
rm -rf dev/
npm install
npm run build
cd /browser-ext/build
export version_string=`grep \"version\" manifest.json | grep -o "[[0-9]*\.]*[0-9]*"`
version_string=$(echo $version_string | tr -d ' ')
zip -r /browser-ext/firefox-sourcegraph-dist-${GIT_HASH:0:8}-${version_string}.xpi *
zip -r /browser-ext/chrome-sourcegraph-dist-${GIT_HASH:0:8}-${version_string}.zip *
