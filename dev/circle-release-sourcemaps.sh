#!/bin/bash
set -ex

VERSION=$1

# release sourcemaps for Splunk JavaScript logging
rm -rf sourcemaps && mkdir -p sourcemaps/assets
cp -R app/node_modules/* sourcemaps/
cp -R app/web_modules/* sourcemaps/
cp -R app/assets/*.map sourcemaps/assets/

tar -czf "sourcemaps-$VERSION.tar.gz" -C sourcemaps .

echo $GCLOUD_SERVICE_ACCOUNT | base64 --decode > gcloud-service-account.json
gcloud auth activate-service-account --key-file gcloud-service-account.json
gsutil cp "sourcemaps-$VERSION.tar.gz" gs://splunk-sourcemaps
