#!/usr/bin/env bash

set -ex

# some agents have a bad cache right now, we can probably remove this in the future
rm -rf $GOPATH/src/github.com/sourcegraph/sourcegraph

if [ ! -d "$GOPATH/src/github.com/sourcegraph/sourcegraph" ]; then
    if [ ! -z "$BUILDKITE_BRANCH" ]; then
        git clone -b "$BUILDKITE_BRANCH" --depth=10 git@github.com:sourcegraph/sourcegraph $GOPATH/src/github.com/sourcegraph/sourcegraph
    else
        git clone --depth=10 git@github.com:sourcegraph/sourcegraph $GOPATH/src/github.com/sourcegraph/sourcegraph
    fi
fi

pushd $GOPATH/src/github.com/sourcegraph/sourcegraph
git fetch
git checkout -f "${OSS_REPO_REVISION:-origin/master}"

if [ ! -z "$BUILDKITE_BRANCH" ] && [ -z "$OSS_REPO_REVISION" ]; then
    # Attempt to check out the branch of the same name in OSS
    echo "Attempting to checkout origin/$BUILDKITE_BRANCH"
    git checkout -f "origin/$BUILDKITE_BRANCH" || git checkout -f "${OSS_REPO_REVISION:-origin/master}"
fi

popd
