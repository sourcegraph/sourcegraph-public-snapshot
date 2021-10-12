#!/usr/bin/env bash

set -e

BASE_URL=http://localhost:3443

echo "--- Runing Lighthouse"
yarn lhci autorun --url="$BASE_URL/" --url="$BASE_URL/search?q=repo:sourcegraph/lighthouse-ci-test-repository+file:index.js" --url="$BASE_URL/github.com/sourcegraph/lighthouse-ci-test-repository" --url="$BASE_URL/github.com/sourcegraph/lighthouse-ci-test-repository/-/blob/index.js"
