#!/bin/bash

cd /browser-ext
rm -rf node_modules
rm -rf build/
rm -rf dev/
npm install
npm run build
cd /browser-ext/build
zip -r /browser-ext/firefox-sourcegraph-dist.xpi *
zip -r /browser-ext/chrome-sourcegraph-dist.zip *
