#!/bin/bash
set -eu -o pipefail

echo "--- check yarn.lock for duplicates"

# Prevent duplicates in yarn.lock/node_modules that lead to errors and bloated bundle sizes

# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --ignore-scripts

echo "Checking for duplicate dependencies in yarn.lock"
yarn run -s yarn-deduplicate --fail --list --strategy fewer ./yarn.lock || {
    echo 'yarn.lock contains duplicate dependencies. Please run `yarn deduplicate` and commit the result.'
    exit 1
}

echo "Checking for duplicate dependencies in cmd/management-console/web/yarn.lock"
yarn run -s yarn-deduplicate --fail --list --strategy fewer ./cmd/management-console/web/yarn.lock || {
    echo 'cmd/management-console/web/yarn.lock contains duplicate dependencies. Please run `yarn deduplicate` and commit the result.'
    exit 1
}
