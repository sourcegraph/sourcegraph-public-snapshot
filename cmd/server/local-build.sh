#!/usr/bin/env bash

echo "Building server locally. This is for testing purposes only."
echo

cd $(dirname "${BASH_SOURCE[0]}")/../..
export GOBIN=$PWD/cmd/server/.bin
set -ex

if [[ -z "${SKIP_PRE_BUILD-}" ]]; then
	./cmd/server/pre-build.sh
fi


# Keep in sync with build.sh
go install -tags dist \
   github.com/sourcegraph/sourcegraph/cmd/server \
   github.com/sourcegraph/sourcegraph/cmd/frontend \
   github.com/sourcegraph/sourcegraph/cmd/github-proxy \
   github.com/sourcegraph/sourcegraph/cmd/gitserver \
   github.com/sourcegraph/sourcegraph/cmd/indexer \
   github.com/sourcegraph/sourcegraph/cmd/query-runner \
   github.com/sourcegraph/sourcegraph/cmd/symbols \
   github.com/sourcegraph/sourcegraph/cmd/repo-updater \
   github.com/sourcegraph/sourcegraph/cmd/searcher \
   github.com/sourcegraph/sourcegraph/cmd/lsp-proxy \
   github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-archive-index \
   github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
   github.com/sourcegraph/sourcegraph/vendor/github.com/google/zoekt/cmd/zoekt-webserver

cat > $GOBIN/syntect_server <<EOF
#!/bin/sh
# Pass through all possible ROCKET env vars
docker run --name=syntect_server --rm -p9238:9238  \
-e QUIET \
-e ROCKET_ENV \
-e ROCKET_ADDRESS \
-e ROCKET_PORT \
-e ROCKET_WORKERS \
-e ROCKET_LOG \
-e ROCKET_SECRET_KEY \
-e ROCKET_LIMITS \
sourcegraph/syntect_server
EOF
chmod +x $GOBIN/syntect_server
