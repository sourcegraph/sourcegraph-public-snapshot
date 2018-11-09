#!/bin/bash

yarn

echo "Building sourcegraph-extension-api..."
yarn workspace sourcegraph run build
echo "Building extensions-client-common..."
yarn workspace @sourcegraph/extensions-client-common run build

echo "Watching the browser extension and dependencies..."
yarn run concurrently --kill-others \
  "yarn workspace sourcegraph run watch:build" \
  "yarn workspace @sourcegraph/extensions-client-common run watch:build" \
  "yarn workspace browser-extensions run dev"
