#!/usr/bin/env bash

set -e

# Setup
yarn global add web-ext
yarn build
rm -rf build/web-ext
mkdir -p build/web-ext

# Sign the bundle
web-ext sign -s build/firefox -a build/web-ext --api-key "$FIREFOX_AMO_ISSUER" --api-secret "$FIREFOX_AMO_SECRET"

# Upload to gcp and make it public
for filename in build/web-ext/*; do
    gsutil cp "build/web-ext/$filename" "gs://sourcegraph-for-firefox/$filename"
    gsutil cp "build/web-ext/$filename" "gs://sourcegraph-for-firefox/latest.xpi"
    gsutil -m acl set -R -a public-read "gs://sourcegraph-for-firefox/$filename"
    gsutil -m acl set -R -a public-read "gs://sourcegraph-for-firefox/latest.xpi"
done

export TS_NODE_COMPILER_OPTIONS="{\"module\":\"commonjs\"}"

gsutil ls gs://sourcegraph-for-firefox | xargs yarn ts-node scripts/build-updates-manifest.ts
gsutil cp src/extension/updates.manifest.json gs://sourcegraph-for-firefox/updates.json
gsutil -m acl set -R -a public-read gs://sourcegraph-for-firefox/updates.json
