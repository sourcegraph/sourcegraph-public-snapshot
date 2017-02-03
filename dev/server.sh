#!/bin/bash

# Runs a Sourcegraph server locally for development. This should still
# be run via `make serve-dev` in the parent directory so that the
# credential env vars get set.

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

GOBIN="$PWD"/vendor/bin go get sourcegraph.com/sourcegraph/sourcegraph/vendor/sourcegraph.com/sqs/rego

export WEBPACK_DEV_SERVER_URL=${WEBPACK_DEV_SERVER_URL:-http://localhost:8080}
export WEBPACK_DEV_SERVER_ADDR=${WEBPACK_DEV_SERVER_ADDR:-127.0.0.1:8080}

function killWebpackDevServer {
	killall -9 webpack-dev-server > /dev/null
}
trap killWebpackDevServer EXIT

curl -Ss -o /dev/null "$WEBPACK_DEV_SERVER_URL" || (cd ui && yarn && yarn run start &)

mkdir -p .bin
env GOBIN=$PWD/.bin go install -v sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,github-proxy,zap}
env SRC_REPOS_DIR=$HOME/.sourcegraph/repos ./.bin/gitserver &
env SRC_GIT_SERVERS=127.0.0.1:3178 LSP_PROXY=127.0.0.1:4388 ./.bin/indexer &
./.bin/github-proxy &
env SRC_GIT_SERVERS=127.0.0.1:3178 ./.bin/zap &

. dev/langservers.lib.bash
detect_dev_langservers

export DEBUG=true
export SRC_APP_DISABLE_SUPPORT_SERVICES=true
export SRC_GIT_SERVERS=127.0.0.1:3178
export GITHUB_BASE_URL=http://127.0.0.1:3180

type ulimit > /dev/null && ulimit -n 10000
exec "$PWD"/vendor/bin/rego \
	 -installenv=GOGC=off,GODEBUG=sbrk=1 \
	 -tags="${GOTAGS-}" \
	 -extra-watches='app/templates/*' \
	 sourcegraph.com/sourcegraph/sourcegraph/cmd/src
