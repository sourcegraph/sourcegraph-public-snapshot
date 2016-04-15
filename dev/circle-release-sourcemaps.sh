#!/bin/bash
set -ex

VERSION=$1

# release sourcemaps for Splunk JavaScript logging
rm -rf sourcemaps && mkdir -p sourcemaps/assets
cp -R app/node_modules/* sourcemaps/
cp -R app/web_modules/* sourcemaps/
cp -R app/assets/*.map sourcemaps/assets/

tar -czf "sourcemaps-$VERSION.tar.gz" -C sourcemaps .
gsutil cp "sourcemaps-$VERSION.tar.gz" gs://splunk-sourcemaps
