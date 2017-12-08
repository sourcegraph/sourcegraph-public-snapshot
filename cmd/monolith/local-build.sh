#!/usr/bin/env bash

echo "Building monolith locally. This is for testing purposes only."
echo

cd $(dirname "${BASH_SOURCE[0]}")/../..
export GOBIN=$PWD/cmd/monolith/.bin
set -ex

if [[ -z "${SKIP_PRE_BUILD-}" ]]; then
	./cmd/monolith/pre-build.sh
fi


# Keep in sync with build.sh
go install -tags dist \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/monolith \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/github-proxy \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/indexer \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-list-updater \
   sourcegraph.com/sourcegraph/sourcegraph/cmd/searcher

cat > $GOBIN/syntect_server <<EOF
#!/bin/sh
# Pass through all possible ROCKET env vars
docker run --name=syntect_server --rm -p3700:80 \
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
