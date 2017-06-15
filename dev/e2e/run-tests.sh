#!/bin/bash
set -e

cd $(dirname "${BASH_SOURCE[0]}")

GOBIN="$PWD"/../../vendor/.bin
env GOBIN=$GOBIN go install -v sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/chromebot

env SRC_APP_URL=http://localhost:3080 ../start.sh &
SERVER_PID=$!
function killServer {
	kill $SERVER_PID
	wait $SERVER_PID
}
trap killServer EXIT

until curl http://localhost:3080/__version; do
	sleep 1
done

node login-test.js | $GOBIN/chromebot
