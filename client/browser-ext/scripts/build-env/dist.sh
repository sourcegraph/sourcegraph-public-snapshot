#!/bin/bash

cd /browser-ext
npm run build
cd /browser-ext/build
zip -r /browser-ext/firefox-sourcegraph-dist.xpi *
zip -r /browser-ext/chrome-sourcegraph-dist.zip *
