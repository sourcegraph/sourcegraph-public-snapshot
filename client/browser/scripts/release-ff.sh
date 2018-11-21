set -e

# Setup
yarn build
rm -rf build/web-ext
mkdir -p build/web-ext

# Sign the bundle
yarn web-ext sign -s build/firefox -a build/web-ext --api-key $FIREFOX_AMO_ISSUER --api-secret $FIREFOX_AMO_SECRET

# Upload to gcp and make it public
for filename in $(ls build/web-ext); do
    gsutil cp build/web-ext/$filename gs://sourcegraph-for-firefox/$filename
    gsutil -m acl set -R -a public-read gs://sourcegraph-for-firefox/$filename
    yarn ts-node scripts/add-version-to-updates-manifest.ts "https://storage.googleapis.com/sourcegraph-for-firefox/$filename"
    echo "https://storage.googleapis.com/sourcegraph-for-firefox/$filename"

    gsutil cp src/extension/updates.manifest.json gs://sourcegraph-for-firefox/updates.json
done

gsutil -m acl set -R -a public-read gs://sourcegraph-for-firefox/updates.json
