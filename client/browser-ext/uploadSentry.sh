#!/bin/bash

set -e

if [ -z "$SENTRY_API_TOKEN" ]; then
    printf "Missing env SENTRY_API_TOKEN; check the infrastructure repository"
    exit 1
fi

version="$(cat chrome/manifest.prod.json | grep "\"version\"" | awk '{ gsub(/"/, "", $2); gsub(/,/, "", $2); print $2 }')"
printf "Creating release $version...\n"
curl https://app.getsentry.com/api/0/projects/sourcegraph/chromefirefox-extension/releases/ -X POST -H "Authorization: Bearer $SENTRY_API_TOKEN" -H "Content-Type: application/json" -d "{\"version\": \"$version\"}"

printf "Uploading artifacts...\n"
curl "https://app.getsentry.com/api/0/projects/sourcegraph/chromefirefox-extension/releases/$version/files/" -X POST -H "Authorization: Bearer $SENTRY_API_TOKEN" -F file=@build/js/inject.bundle.js -F name="chrome-extension://ndcgbikhadhkobldlipfhjjkdbjfonpk/js/inject.bundle.js"
printf "\n"
curl "https://app.getsentry.com/api/0/projects/sourcegraph/chromefirefox-extension/releases/$version/files/" -X POST -H "Authorization: Bearer $SENTRY_API_TOKEN" -F file=@build/js/inject.bundle.js.map -F name="inject.bundle.js.map"
