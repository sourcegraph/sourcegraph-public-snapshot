#!/bin/bash
set -e

cd $(dirname "${BASH_SOURCE[0]}")

if [ ! -f "./$1" ]; then
	echo "usage: run-test.sh [.test.js file]"
	exit 1
fi

GOBIN="$PWD"/../../vendor/.bin
env GOBIN=$GOBIN go install -v sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/chromebot

env SRC_APP_URL=http://localhost:3080 SHELL=/bin/bash ../dev/start.sh &
SERVER_PID=$!
function killServer {
	kill $SERVER_PID
	wait $SERVER_PID
}
trap killServer EXIT

until [ "$(curl -s -w "%{http_code}" http://localhost:3080/.assets/scripts/app.bundle.js -o /dev/null)" == "200" ]; do
	sleep 1
done

node $1 | $GOBIN/chromebot
