#!/usr/bin/env bash

set -ex

# some agents have a bad cache right now, we can probably remove this in the future
rm -rf $GOPATH/src/github.com/sourcegraph/sourcegraph

if [ ! -d "$GOPATH/src/github.com/sourcegraph/sourcegraph" ]; then
  git clone --depth=10 git@github.com:sourcegraph/sourcegraph $GOPATH/src/github.com/sourcegraph/sourcegraph
fi

pushd $GOPATH/src/github.com/sourcegraph/sourcegraph
git fetch
git checkout -f "${OSS_REPO_REVISION:-origin/master}"
popd
