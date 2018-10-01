#!/usr/bin/env bash
#
# Updates go.mod when running in CI to remove the replace directive

set -e

if [ ! -z "$(git status --porcelain)" ]; then
    echo "working directory not clean, aborting"
    exit 1
fi

GO111MODULE=on go mod edit -dropreplace=github.com/sourcegraph/sourcegraph
if [ ! -z "$(git status --porcelain)" ]; then
    if [ -z "$(git config user.email)" ]; then
        git config --global user.email "no-reply@sourcegraph.com"
    fi
    if [ -z "$(git config user.name)" ]; then
        git config --global user.name "Mr. Continuous Integration"
    fi
    git commit -a -m'!!! CI-ONLY (DO NOT MERGE THIS COMMIT): remove replace directive in go.mod'
fi
